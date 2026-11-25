"""Split the original Maple UI font without changing glyphs or font-display.

Requires Python 3.12+, fonttools==4.68.0 and brotli==1.2.1.
Generated WOFF2 files are committed; production builds do not need Python.
"""

import hashlib
import io
import json
import logging
import sys
import argparse
from collections import Counter
from importlib.metadata import PackageNotFoundError, version
from pathlib import Path

from fontTools import subset
from fontTools.ttLib import TTFont


PROJECT_ROOT = Path(__file__).resolve().parent.parent
SOURCE_FONT_PATH = PROJECT_ROOT / "scripts/font-source-assets/MapleUI-Original-Complete-CJK-And-Extended.f8ce6d72e8cb.woff2"
PUBLIC_FONT_DIRECTORY = PROJECT_ROOT / "public/fonts"
LATIN_FONT_PATH = PUBLIC_FONT_DIRECTORY / "MapleUI-Latin.00af3fec5b34.woff2"
MANIFEST_PATH = PROJECT_ROOT / "config/storefront/storefront-cjk-and-extended-font-character-shards.manifest.json"
PRIMARY_INTERFACE_NAMESPACES = {
    "header", "common", "dockMenu", "cookieConsent", "home", "footer",
    "nav", "menu", "shopCategoryMenu", "productCategoryNavigationCards",
    "storefrontDataNotice", "faq",
}
MAXIMUM_CHARACTERS_PER_SHARD = 512
MAXIMUM_FONT_BYTES_PER_SHARD = 128 * 1024
GENERATOR_DEPENDENCIES = {"fonttools": "4.68.0", "brotli": "1.2.1"}


def verify_pinned_generator_dependencies():
    if sys.version_info < (3, 12):
        raise RuntimeError(f"Font shard generation requires Python 3.12 or newer; found {sys.version.split()[0]}.")
    for package_name, expected_version in GENERATOR_DEPENDENCIES.items():
        try:
            installed_version = version(package_name)
        except PackageNotFoundError as error:
            raise RuntimeError(
                f"Missing generator dependency {package_name}=={expected_version}; "
                "install scripts/requirements-storefront-font-shard-generation.txt."
            ) from error
        if installed_version != expected_version:
            raise RuntimeError(
                f"Generator requires {package_name}=={expected_version}, "
                f"but {installed_version} is installed."
            )


def collect_message_strings(value):
    if isinstance(value, str):
        yield value
    elif isinstance(value, dict):
        for entry in value.values():
            yield from collect_message_strings(entry)
    elif isinstance(value, list):
        for entry in value:
            yield from collect_message_strings(entry)


def serialize_exact_unicode_ranges(code_points):
    runs = []
    for code_point in sorted(code_points):
        if runs and runs[-1][1] + 1 == code_point:
            runs[-1][1] = code_point
        else:
            runs.append([code_point, code_point])
    return ", ".join(
        f"U+{start:04X}" if start == end else f"U+{start:04X}-{end:04X}"
        for start, end in runs
    )


def classify_original_character_script(code_point):
    if 0x3400 <= code_point <= 0x9FFF or 0xF900 <= code_point <= 0xFAFF or code_point >= 0x20000:
        return "Han-Ideographs"
    if 0xAC00 <= code_point <= 0xD7AF:
        return "Korean-Hangul"
    if 0x2E80 <= code_point <= 0x33FF or 0xFE30 <= code_point <= 0xFE4F or 0xFF00 <= code_point <= 0xFFEF:
        return "Kana-And-CJK-Punctuation"
    if 0x0370 <= code_point <= 0x052F or 0x1F00 <= code_point <= 0x1FFF:
        return "Greek-And-Cyrillic"
    return "Phonetic-Armenian-Georgian-And-Extended-Symbols"


def create_unchanged_glyph_font_subset(source_bytes, code_points):
    font = TTFont(io.BytesIO(source_bytes), recalcTimestamp=False)
    options = subset.Options()
    options.layout_features = ["*"]
    options.name_IDs = ["*"]
    options.name_legacy = True
    options.name_languages = ["*"]
    options.recalc_timestamp = False
    options.recalc_bounds = False
    options.hinting = True
    options.notdef_glyph = True
    options.notdef_outline = True
    subsetter = subset.Subsetter(options=options)
    subsetter.populate(unicodes=code_points)
    subsetter.subset(font)
    font.flavor = "woff2"
    output = io.BytesIO()
    font.save(output, reorderTables=True)
    return output.getvalue()


def generate_storefront_font_character_shards():
    arguments = argparse.ArgumentParser(description=__doc__)
    arguments.add_argument('--decoded-source-ttf', type=Path, help='Optional predecoded original TTF to avoid repeated WOFF2 reconstruction; glyph parity is checked separately.')
    options = arguments.parse_args()
    verify_pinned_generator_dependencies()
    logging.getLogger("fontTools.subset").setLevel(logging.ERROR)
    source_bytes = SOURCE_FONT_PATH.read_bytes()
    if options.decoded_source_ttf:
        subset_source_bytes = options.decoded_source_ttf.read_bytes()
    else:
        original_font = TTFont(io.BytesIO(source_bytes), recalcTimestamp=False)
        original_font.flavor = None
        decoded_output = io.BytesIO()
        original_font.save(decoded_output)
        subset_source_bytes = decoded_output.getvalue()
    original_font = TTFont(io.BytesIO(subset_source_bytes))
    if any(
        table.platformID == 0 and table.platEncID == 5 and table.format == 14
        for table in original_font["cmap"].tables
    ):
        raise ValueError(
            "The source font contains cmap format 14 Unicode variation sequences; "
            "extend the shard generator before regenerating."
        )
    source_code_points = set(original_font.getBestCmap())
    latin_font = TTFont(LATIN_FONT_PATH)
    latin_code_points = set(latin_font.getBestCmap())
    original_code_points = source_code_points - latin_code_points
    primary_frequency = Counter()
    for source_path in sorted((PROJECT_ROOT / "app/i18n/messages").glob("*/*.json")):
        for message in collect_message_strings(json.loads(source_path.read_text(encoding="utf-8"))):
            characters = [ord(character) for character in message if ord(character) in original_code_points]
            if source_path.stem in PRIMARY_INTERFACE_NAMESPACES:
                primary_frequency.update(characters)

    character_groups = {}
    for code_point in sorted(original_code_points):
        group = (
            classify_original_character_script(code_point),
            "Common-Storefront-Interface" if code_point in primary_frequency else "Remaining-Complete-Coverage",
        )
        character_groups.setdefault(group, []).append(code_point)

    # Read the previous inventory before writing; only its exact retired outputs may be removed.
    previous_manifest = json.loads(MANIFEST_PATH.read_text(encoding="utf-8")) if MANIFEST_PATH.exists() else {"shards": []}
    shards = []
    for (script, role), code_points in sorted(character_groups.items()):
        if role == "Common-Storefront-Interface":
            code_points.sort(key=lambda code_point: (-primary_frequency[code_point], code_point))
        else:
            # Contiguous rare-character groups keep CSS range declarations compact.
            code_points.sort()
        pending_groups = [code_points[index:index + MAXIMUM_CHARACTERS_PER_SHARD] for index in range(0, len(code_points), MAXIMUM_CHARACTERS_PER_SHARD)]
        group_index = 0
        while pending_groups:
            selected_code_points = pending_groups.pop(0)
            font_bytes = create_unchanged_glyph_font_subset(subset_source_bytes, selected_code_points)
            if len(font_bytes) > MAXIMUM_FONT_BYTES_PER_SHARD and len(selected_code_points) > 1:
                midpoint = len(selected_code_points) // 2
                pending_groups[0:0] = [selected_code_points[:midpoint], selected_code_points[midpoint:]]
                continue
            group_index += 1
            digest = hashlib.sha256(font_bytes).hexdigest()
            filename = f"MapleUI-CJK-{script}-{role}-Character-Group-{group_index:03d}.{digest[:12]}.woff2"
            (PUBLIC_FONT_DIRECTORY / filename).write_bytes(font_bytes)
            shards.append({
                "fontFamily": "MapleUICJK", "filename": filename, "script": script,
                "role": role, "unicodeRange": serialize_exact_unicode_ranges(selected_code_points),
                "bytes": len(font_bytes), "sha256": digest, "codePointCount": len(selected_code_points),
            })
            print(f"{script} / {role} / {group_index}: {len(selected_code_points)} characters, {len(font_bytes)} bytes", flush=True)

    manifest = {
        "schemaVersion": 2,
        "sourceFontPath": SOURCE_FONT_PATH.relative_to(PROJECT_ROOT).as_posix(),
        "sourceSha256": hashlib.sha256(source_bytes).hexdigest(),
        "sourceBytes": len(source_bytes),
        "sourceCodePointCount": len(source_code_points),
        "latinExcludedCodePointCount": len(source_code_points & latin_code_points),
        "coveredCodePointCount": len(original_code_points),
        "maximumFontBytesPerShard": MAXIMUM_FONT_BYTES_PER_SHARD,
        "generatorDependencies": GENERATOR_DEPENDENCIES,
        "shards": shards,
    }
    MANIFEST_PATH.write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    font_faces = "\n\n".join(
        "@font-face {\n  font-family: 'MapleUICJK';\n"
        f"  src: url('/fonts/{shard['filename']}') format('woff2');\n"
        "  font-weight: 100 900;\n  font-style: normal;\n  font-display: block;\n"
        f"  unicode-range: {shard['unicodeRange']};\n}}"
        for shard in shards
    )
    font_css_path = PROJECT_ROOT / "app/assets/css/tailwind.css"
    font_css = font_css_path.read_text(encoding="utf-8")
    start_marker = "/* BEGIN GENERATED MAPLE UI CHARACTER SHARDS */"
    end_marker = "/* END GENERATED MAPLE UI CHARACTER SHARDS */"
    generated_css = f"{start_marker}\n{font_faces}\n{end_marker}"
    if start_marker in font_css:
        start = font_css.index(start_marker)
        end = font_css.index(end_marker) + len(end_marker)
    else:
        start = font_css.index("@font-face {\n  font-family: 'MapleUICJK';")
        end = font_css.index("\n}", start) + 2
    font_css = font_css[:start] + generated_css + font_css[end:]
    font_css_path.write_text(font_css, encoding="utf-8", newline="\n")
    # Stripe uses the same family, files, ranges and loading policy as the storefront.
    stripe_font_css = font_css[font_css.index("@font-face"):font_css.index("@layer base")].strip()
    (PUBLIC_FONT_DIRECTORY / "maple-ui.css").write_text(stripe_font_css + "\n", encoding="utf-8", newline="\n")
    retained_filenames = {shard["filename"] for shard in shards}
    for shard in previous_manifest["shards"]:
        if shard["filename"] not in retained_filenames:
            retired_path = PUBLIC_FONT_DIRECTORY / shard["filename"]
            if retired_path.parent.resolve() != PUBLIC_FONT_DIRECTORY.resolve():
                raise ValueError("Retired font output escapes the public font directory")
            retired_path.unlink(missing_ok=True)
    print(f"Generated {len(shards)} disjoint shards covering all {len(original_code_points)} original cmap characters not covered by the Latin face.")


if __name__ == "__main__":
    generate_storefront_font_character_shards()
