import fs from 'node:fs'
import path from 'node:path'
import process from 'node:process'
import { fileURLToPath } from 'node:url'

const scriptDir = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(scriptDir, '..')
const sourcePath = path.join(repoRoot, 'shared', 'website-name-defaults.json')

const outputPaths = {
  go: path.join(
    repoRoot,
    'go-backend',
    'internal',
    'domain',
    'setting',
    'website_name_defaults_generated.go',
  ),
  nuxt: path.join(
    repoRoot,
    'nuxt-i18n',
    'app',
    'data',
    'websiteNameDefaults.generated.ts',
  ),
}

const expectedLocales = ['en', 'zh_cn']
const expectedFields = ['status', 'intro', 'eyebrow', 'title', 'body', 'note']

const historicalMigrationChecks = [
  {
    path: path.join(repoRoot, 'go-backend', 'migrations', '202_seed_website_name_defaults.up.sql'),
    required: (source) => expectedFields.flatMap((fieldName) => [
      sourceFieldKey(source, fieldName),
      fieldName,
    ]),
  },
  {
    path: path.join(repoRoot, 'go-backend', 'migrations', '202_seed_website_name_defaults.down.sql'),
    required: (source) => expectedFields.map((fieldName) => sourceFieldKey(source, fieldName)),
  },
  {
    path: path.join(repoRoot, 'go-backend', 'migrations', '205_remove_website_name_placeholders.up.sql'),
    required: (source) => ['body', 'note'].flatMap((fieldName) => [
      sourceFieldKey(source, fieldName),
      fieldName,
    ]),
  },
  {
    path: path.join(repoRoot, 'go-backend', 'migrations', '207_namespace_website_name_setting_keys.up.sql'),
    required: (source) => expectedFields.flatMap((fieldName) => [
      sourceFieldKey(source, fieldName),
      fieldName,
    ]),
  },
  {
    path: path.join(repoRoot, 'go-backend', 'migrations', '207_namespace_website_name_setting_keys.down.sql'),
    required: (source) => expectedFields.flatMap((fieldName) => [
      sourceFieldKey(source, fieldName),
      fieldName,
    ]),
  },
]

function sourceFieldKey(source, fieldName) {
  return source.fields[fieldName].key
}

function fail(message) {
  throw new Error(`Invalid ${path.relative(repoRoot, sourcePath)}: ${message}`)
}

function isPlainObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

function assertString(value, label) {
  if (typeof value !== 'string') {
    fail(`${label} must be a string`)
  }
}

function validateSource(source) {
  if (!isPlainObject(source)) fail('root must be an object')
  assertString(source.group, 'group')
  if (source.group !== 'website_name') fail('group must be website_name')
  if (!isPlainObject(source.fields)) fail('fields must be an object')
  if (!isPlainObject(source.locales)) fail('locales must be an object')

  const fieldNames = Object.keys(source.fields)
  if (JSON.stringify(fieldNames) !== JSON.stringify(expectedFields)) {
    fail(`fields must be exactly: ${expectedFields.join(', ')}`)
  }

  for (const fieldName of expectedFields) {
    const field = source.fields[fieldName]
    if (!isPlainObject(field)) fail(`fields.${fieldName} must be an object`)
    assertString(field.key, `fields.${fieldName}.key`)
    assertString(field.description, `fields.${fieldName}.description`)
    if (!field.key.startsWith('website_name_')) {
      fail(`fields.${fieldName}.key must start with website_name_`)
    }
  }
  const fieldKeys = expectedFields.map((fieldName) => source.fields[fieldName].key)
  if (new Set(fieldKeys).size !== fieldKeys.length) {
    fail('fields keys must be unique')
  }

  const localeNames = Object.keys(source.locales)
  if (JSON.stringify(localeNames) !== JSON.stringify(expectedLocales)) {
    fail(`locales must be exactly: ${expectedLocales.join(', ')}`)
  }

  for (const locale of expectedLocales) {
    const values = source.locales[locale]
    if (!isPlainObject(values)) fail(`locales.${locale} must be an object`)
    const valueNames = Object.keys(values)
    if (JSON.stringify(valueNames) !== JSON.stringify(expectedFields)) {
      fail(`locales.${locale} must contain exactly: ${expectedFields.join(', ')}`)
    }
    for (const fieldName of expectedFields) {
      assertString(values[fieldName], `locales.${locale}.${fieldName}`)
    }
  }
}

function validateHistoricalMigrations(source) {
  for (const migration of historicalMigrationChecks) {
    if (!fs.existsSync(migration.path)) {
      fail(`historical migration is missing: ${path.relative(repoRoot, migration.path)}`)
    }

    const contents = fs.readFileSync(migration.path, 'utf8')
    for (const fragment of migration.required(source)) {
      if (!contents.includes(fragment)) {
        fail(`historical migration ${path.relative(repoRoot, migration.path)} is missing ${fragment}`)
      }
    }

    if (migration.path.endsWith('202_seed_website_name_defaults.up.sql')) {
      for (const locale of expectedLocales) {
        for (const fieldName of expectedFields) {
          const defaultValue = source.locales[locale][fieldName]
          if (defaultValue && contents.includes(defaultValue)) {
            fail('202 seed migration must not contain editable default prose')
          }
        }
      }
    }

    if (migration.path.endsWith('202_seed_website_name_defaults.down.sql') && !contents.includes('AND value = \'\'')) {
      fail('202 seed migration rollback must preserve non-empty administrator content')
    }

    if (migration.path.endsWith('207_namespace_website_name_setting_keys.up.sql')) {
      for (const fragment of [
        'UPDATE settings AS namespaced',
        "NULLIF(BTRIM(namespaced.value), '') IS NULL",
        "NULLIF(BTRIM(legacy.value), '') IS NOT NULL",
      ]) {
        if (!contents.includes(fragment)) {
          fail(`207 namespace migration is missing ${fragment}`)
        }
      }
    }
  }
}

function goString(value) {
  return JSON.stringify(value)
}

function goFieldName(fieldName) {
  return fieldName
    .split('_')
    .map((part) => `${part.slice(0, 1).toUpperCase()}${part.slice(1)}`)
    .join('')
}

const generatedGoFieldNames = ['Locale', ...expectedFields.map(goFieldName)]
const generatedGoFieldNameWidth = Math.max(...generatedGoFieldNames.map((name) => name.length))
const generatedGoKeyNames = [
  'WebsiteNameGroup',
  ...expectedFields.map((fieldName) => `WebsiteNameKey${goFieldName(fieldName)}`),
]
const generatedGoKeyNameWidth = Math.max(...generatedGoKeyNames.map((name) => name.length))

function generatedHeader(sourceDescription) {
  return `// Code generated by scripts/generate-website-name-defaults.mjs from ${sourceDescription}. DO NOT EDIT.
`
}

function renderGo(source) {
  const lines = [
    generatedHeader('shared/website-name-defaults.json').trimEnd(),
    'package setting',
    '',
    'const (',
  ]

  const groupPadding = ' '.repeat(generatedGoKeyNameWidth - 'WebsiteNameGroup'.length + 1)
  lines.push(`\tWebsiteNameGroup${groupPadding}= ${goString(source.group)}`)

  for (const fieldName of expectedFields) {
    const generatedKeyName = `WebsiteNameKey${goFieldName(fieldName)}`
    const padding = ' '.repeat(generatedGoKeyNameWidth - generatedKeyName.length + 1)
    lines.push(`\t${generatedKeyName}${padding}= ${goString(source.fields[fieldName].key)}`)
  }

  lines.push(')', '')
  lines.push('func WebsiteNameSettingDescription(key string) string {')
  lines.push('\tswitch key {')
  for (const fieldName of expectedFields) {
    const generatedKeyName = `WebsiteNameKey${goFieldName(fieldName)}`
    lines.push(`\tcase ${generatedKeyName}:`)
    lines.push(`\t\treturn ${goString(`Why this name page: ${source.fields[fieldName].description}`)}`)
  }
  lines.push('\tdefault:')
  lines.push('\t\treturn key')
  lines.push('\t}')
  lines.push('}', '')
  lines.push(
    'var websiteNameDefaultSettingsByLocale = map[string]WebsiteNameSettings{',
  )

  for (const locale of expectedLocales) {
    lines.push(`\t${goString(locale)}: {`)
    const localePadding = ' '.repeat(generatedGoFieldNameWidth - 'Locale'.length + 1)
    lines.push(`\t\tLocale:${localePadding}${goString(locale)},`)
    for (const fieldName of expectedFields) {
      const generatedFieldName = goFieldName(fieldName)
      const padding = ' '.repeat(generatedGoFieldNameWidth - generatedFieldName.length + 1)
      lines.push(`\t\t${generatedFieldName}:${padding}${goString(source.locales[locale][fieldName])},`)
    }
    lines.push('\t},')
  }

  lines.push('}')
  lines.push('')
  return `${lines.join('\n')}`
}

function renderNuxt(source) {
  const lines = [
    '// Code generated by scripts/generate-website-name-defaults.mjs from shared/website-name-defaults.json. DO NOT EDIT.',
    '',
    'export interface WebsiteNameDefaultValues {',
  ]

  for (const fieldName of expectedFields) {
    lines.push(`  ${fieldName}: string`)
  }

  lines.push('}', '')
  lines.push(`export type WebsiteNameDefaultLocale = ${expectedLocales.map((locale) => `'${locale}'`).join(' | ')}`)
  lines.push('')
  lines.push('export const WEBSITE_NAME_DEFAULTS: Record<WebsiteNameDefaultLocale, WebsiteNameDefaultValues> = {')

  for (const locale of expectedLocales) {
    lines.push(`  ${locale}: {`)
    for (const fieldName of expectedFields) {
      lines.push(`    ${fieldName}: ${JSON.stringify(source.locales[locale][fieldName])},`)
    }
    lines.push('  },')
  }

  lines.push('}', '')
  return lines.join('\n')
}

function renderOutputs(source) {
  return {
    [outputPaths.go]: renderGo(source),
    [outputPaths.nuxt]: renderNuxt(source),
  }
}

function writeOrCheck(outputs, checkOnly) {
  const staleFiles = []

  for (const [filePath, content] of Object.entries(outputs)) {
    const existing = fs.existsSync(filePath) ? fs.readFileSync(filePath, 'utf8') : null
    if (existing === content) continue

    staleFiles.push(path.relative(repoRoot, filePath).replaceAll(path.sep, '/'))
    if (!checkOnly) {
      fs.mkdirSync(path.dirname(filePath), { recursive: true })
      fs.writeFileSync(filePath, content, 'utf8')
    }
  }

  if (checkOnly && staleFiles.length > 0) {
    throw new Error(`Generated files are stale:\n${staleFiles.map((file) => `- ${file}`).join('\n')}`)
  }

  return staleFiles
}

const source = JSON.parse(fs.readFileSync(sourcePath, 'utf8'))
validateSource(source)
validateHistoricalMigrations(source)

const changedFiles = writeOrCheck(renderOutputs(source), process.argv.includes('--check'))

if (process.argv.includes('--check')) {
  console.log('Website name defaults are up to date.')
} else if (changedFiles.length > 0) {
  console.log(`Generated ${changedFiles.length} website name default files.`)
} else {
  console.log('Website name default files were already up to date.')
}
