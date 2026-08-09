declare module 'ttf2woff' {
  const convert: (input: Buffer) => { buffer: ArrayBuffer }

  export default convert
}
