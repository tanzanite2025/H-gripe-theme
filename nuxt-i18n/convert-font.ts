import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

type Ttf2Woff2Module = {
  default: (input: Buffer) => Uint8Array | Buffer
}

type Ttf2WoffModule = {
  default: (input: Buffer) => { buffer: ArrayBuffer }
}

const scriptDir = path.dirname(fileURLToPath(import.meta.url))

const inputPath = path.join(scriptDir, 'app/assets/fonts/AerialFasterRegular-Yqd5o.ttf')
const outputWoff2Path = path.join(scriptDir, 'app/assets/fonts/AerialFasterRegular.woff2')
const outputWoffPath = path.join(scriptDir, 'app/assets/fonts/AerialFasterRegular.woff')

async function main(): Promise<void> {
  console.log('Starting font conversion...')
  console.log('Input file:', inputPath)

  if (!fs.existsSync(inputPath)) {
    console.error('Error: font file not found.')
    console.error('Check the path:', inputPath)
    process.exit(1)
  }

  const ttfBuffer = fs.readFileSync(inputPath)
  console.log('Loaded TTF file, size:', `${(ttfBuffer.length / 1024).toFixed(2)} KB`)

  try {
    const ttf2woff2Module = await import('ttf2woff2') as Ttf2Woff2Module
    const woff2Buffer = Buffer.from(ttf2woff2Module.default(ttfBuffer))
    fs.writeFileSync(outputWoff2Path, woff2Buffer)
    console.log('Generated WOFF2, size:', `${(woff2Buffer.length / 1024).toFixed(2)} KB`)
    console.log('Compression ratio:', `${((1 - woff2Buffer.length / ttfBuffer.length) * 100).toFixed(1)}%`)
  } catch (error: unknown) {
    console.error('WOFF2 conversion failed:', error instanceof Error ? error.message : error)
    console.log('Install dependency first: npm install ttf2woff2')
  }

  try {
    const ttf2woffModule = await import('ttf2woff') as Ttf2WoffModule
    const woffBuffer = Buffer.from(ttf2woffModule.default(ttfBuffer).buffer)
    fs.writeFileSync(outputWoffPath, woffBuffer)
    console.log('Generated WOFF, size:', `${(woffBuffer.length / 1024).toFixed(2)} KB`)
    console.log('Compression ratio:', `${((1 - woffBuffer.length / ttfBuffer.length) * 100).toFixed(1)}%`)
  } catch (error: unknown) {
    console.error('WOFF conversion failed:', error instanceof Error ? error.message : error)
    console.log('Install dependency first: npm install ttf2woff')
  }

  console.log('\nFont conversion complete.')
  console.log('Output files:')
  console.log('  - WOFF2:', outputWoff2Path)
  console.log('  - WOFF:', outputWoffPath)
}

main().catch((error: unknown) => {
  console.error(error instanceof Error ? error.message : error)
  process.exit(1)
})
