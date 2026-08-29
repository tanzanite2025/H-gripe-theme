# Spoke Tension Ratio Architecture

## Decision

The calculator treats the tension ratio as a derived geometry result, not as
the 2:1 or 1:1 spoke-hole topology ratio.

- `2:1` / `1:1`: spoke count distribution between the two sides.
- `T_left / T_right`: estimated per-spoke static tension ratio required by the
  two bracing angles.
- `lowerToHigher`: normalized ratio in the range `(0, 1]`, suitable for
  display as `78%`.

The report's `78%` reference is consistent with `20.3 / 26.2 = 77.5%`.
That means `WL` and `WR` must be treated as effective bracing distances for
that reference. The report must not apply the `2.8 mm` offset a second time
unless those values are explicitly defined as hub-center distances.

## Geometry Convention

`rimOffsetMm` is optional and uses this sign convention:

- `0`: the rim center is on the hub center plane.
- Positive: the rim center moves toward the right flange.
- Negative: the rim center moves toward the left flange.

With `wLeft` and `wRight` stored as positive hub-center-to-flange distances:

```text
dLeft  = abs(wLeft  + rimOffsetMm)
dRight = abs(wRight - rimOffsetMm)
sin(thetaLeft)  = dLeft  / LLeft
sin(thetaRight) = dRight / LRight
TLeft / TRight   = sin(thetaRight) / sin(thetaLeft)
```

`LLeft` and `LRight` are the calculated or verified spoke lengths for the
same geometry. The result also exposes both directional ratios, the
lower-tension side, the normalized lower-to-higher ratio, and both bracing
angles.

The current catalog fields `leftFlange` and `rightFlange` are interpreted as
hub-center-to-flange distances. For the report's `78%` reference, the values
`26.2 mm` and `20.3 mm` reproduce `20.3 / 26.2 = 77.5%` only when they are
used as effective bracing distances, so the calculation must use
`rimOffsetMm = 0` for that reference. If they are raw hub-center distances,
`δ = 2.8 mm` is applied exactly once and the resulting ratio is a different
geometric prediction.

## Integration Points

1. Go service input accepts `rimOffsetMm`; the calculation result returns
   `tensionRatio`.
2. The public API preserves the existing length fields and adds the new
   fields, so older clients remain compatible.
3. The Nuxt calculator exposes rim-center offset for each wheel and renders
   the predicted ratio beside the left/right spoke lengths.
4. A non-zero manual offset disables legacy verified lengths that do not carry
   the same offset geometry; the calculator falls back to the calculated
   lengths.

## Catalog Evolution

The current catalog does not persist a wheel-build dish/offset contract, so
`rimOffsetMm` is intentionally a calculation input for this phase. Once the
report's sign convention is confirmed with measured builds, add these optional
fields to `spoke_build_presets`:

- `rim_offset_mm`
- `target_tension_ratio_percent`
- `tension_ratio_notes`

The target is a validation/reference value. It must not replace the calculated
ratio, and it must not be presented as a measured assembly result.

## Verification

For every published preset, the acceptance check should compare:

- calculated spoke lengths against the verified cut lengths;
- calculated `lowerToHigher` against the target ratio, if one exists;
- measured tension ratio after assembly, when available;
- the allowed rim offset and minimum bracing-angle/clearance rules from the
  engineering report.
