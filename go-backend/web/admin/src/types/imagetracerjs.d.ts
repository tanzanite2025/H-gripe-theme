declare module 'imagetracerjs' {
  export interface ImageTracerOptions {
    blurradius?: number
    blurdelta?: number
    colorsampling?: number
    colorquantcycles?: number
    desc?: boolean
    layering?: number
    linefilter?: boolean
    ltres?: number
    mincolorratio?: number
    numberofcolors?: number
    pathomit?: number
    qtres?: number
    rightangleenhance?: boolean
    roundcoords?: number
    scale?: number
    strokewidth?: number
    viewbox?: boolean
    pal?: Array<{ r: number; g: number; b: number; a: number }>
  }

  interface ImageTracerApi {
    imagedataToSVG(
      imageData: ImageData,
      options?: ImageTracerOptions | string,
    ): string
  }

  const imageTracer: ImageTracerApi

  export default imageTracer
}
