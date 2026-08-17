import { access, cp, readFile } from 'node:fs/promises'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { familySync } from 'detect-libc'

const scriptDirectory = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(scriptDirectory, '../..')
const sourceNodeModules = join(projectRoot, 'node_modules')
const outputServerDirectory = join(projectRoot, '.output', 'server')
const outputNodeModules = join(outputServerDirectory, 'node_modules')

const exists = async (path: string): Promise<boolean> => {
  try {
    await access(path)
    return true
  } catch {
    return false
  }
}

const runtimePlatform = (): string => {
  if (process.platform !== 'linux') {
    return `${process.platform}-${process.arch}`
  }

  return `${familySync() === 'musl' ? 'linuxmusl' : 'linux'}-${process.arch}`
}

const copyRuntimePackage = async (packageName: string): Promise<string> => {
  const source = join(sourceNodeModules, ...packageName.split('/'))
  const destination = join(outputNodeModules, ...packageName.split('/'))

  if (!(await exists(source))) {
    throw new Error(`Missing installed sharp runtime package: ${packageName}`)
  }

  await cp(source, destination, { recursive: true, force: true })
  return destination
}

const platform = runtimePlatform()
const nativePackageName = `@img/sharp-${platform}`
const libvipsPackageName = platform.startsWith('linux')
  ? `@img/sharp-libvips-${platform}`
  : null

if (!(await exists(outputServerDirectory))) {
  throw new Error(`Missing Nitro server output: ${outputServerDirectory}`)
}

const copiedPackages = [await copyRuntimePackage(nativePackageName)]
if (libvipsPackageName) {
  copiedPackages.push(await copyRuntimePackage(libvipsPackageName))
}

const nativePackageDirectory = join(
  outputNodeModules,
  ...nativePackageName.split('/'),
)
const nativePackage = JSON.parse(
  await readFile(join(nativePackageDirectory, 'package.json'), 'utf8'),
) as { version?: string }
const nativeBinary = join(
  nativePackageDirectory,
  'lib',
  `sharp-${platform}-${nativePackage.version || 'unknown'}.node`,
)

if (!(await exists(nativeBinary))) {
  throw new Error(`Sharp native binary was not copied to Nitro output: ${nativeBinary}`)
}

if (libvipsPackageName) {
  const libvipsPackageDirectory = join(
    outputNodeModules,
    ...libvipsPackageName.split('/'),
  )
  const libvipsFiles = await import('node:fs/promises').then(({ readdir }) =>
    readdir(join(libvipsPackageDirectory, 'lib')),
  )

  if (!libvipsFiles.some(file => file.startsWith('libvips-cpp'))) {
    throw new Error(`Sharp libvips runtime was not copied to Nitro output: ${libvipsPackageDirectory}`)
  }
}

const outputSharpEntry = join(outputNodeModules, 'sharp', 'dist', 'index.mjs')
if (!(await exists(outputSharpEntry))) {
  throw new Error(`Sharp server entry is missing from Nitro output: ${outputSharpEntry}`)
}

await import(pathToFileURL(outputSharpEntry).href)

console.log(
  `[sharp-runtime] OK: ${platform}; copied ${copiedPackages.length} native package(s) and loaded sharp from .output/server.`,
)
