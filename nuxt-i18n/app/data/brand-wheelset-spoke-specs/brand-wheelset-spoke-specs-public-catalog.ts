export type PublicBrandWheelsetPublicationStatus = 'draft' | 'published'
export type PublicBrandWheelsetLifecycleStatus = 'current' | 'legacy'
export type PublicBrandWheelPosition = 'front' | 'rear'
export type PublicBrandWheelSpokeSide = 'left' | 'right'
export type PublicBrandSpokeHeadType = 'straight-pull' | 'j-bend' | 'unknown'

export interface PublicBrandWheelSideSpokeSpec {
  side: PublicBrandWheelSpokeSide
  lengthMm: number | null
  spokeModel: string
  headType: PublicBrandSpokeHeadType
}

export interface PublicBrandWheelPositionSpokeSpec {
  position: PublicBrandWheelPosition
  spokeCount: number
  lacingPattern: string
  sides: PublicBrandWheelSideSpokeSpec[]
}

export interface PublicBrandWheelsetRecord {
  slug: string
  model: string
  lifecycleStatus: PublicBrandWheelsetLifecycleStatus
  nippleModel: string
  nippleLengthMm: number | null
  rim: {
    depthMm: number
    depthFrontMm?: number
    depthRearMm?: number
  }
  wheels: PublicBrandWheelPositionSpokeSpec[]
}

export interface PublicBrandWheelsetCatalog {
  brandSlug: string
  brandName: string
  publicationStatus: PublicBrandWheelsetPublicationStatus
  sourceCheckedAt: string | null
  wheelsets: PublicBrandWheelsetRecord[]
}

export const publicBrandWheelsetSpokeCatalogs: PublicBrandWheelsetCatalog[] = [
  {
    "brandSlug": "dt-swiss",
    "brandName": "DT Swiss",
    "publicationStatus": "published",
    "sourceCheckedAt": "2026-10-01",
    "wheelsets": [
      {
        "slug": "arc-1100-dicut-db-38",
        "model": "ARC 1100 DICUT DB 38",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 38
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 285,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 285,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1100-dicut-db-55",
        "model": "ARC 1100 DICUT DB 55",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 55
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 269,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 271,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 266,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 263,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1100-dicut-db-65",
        "model": "ARC 1100 DICUT DB 65",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 65
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 259,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 261,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 255,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 253,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1100-dicut-db-85",
        "model": "ARC 1100 DICUT DB 85",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 85
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 239,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 241,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 236,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 233,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1400-dicut-db-38",
        "model": "ARC 1400 DICUT DB 38",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 38
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1400-dicut-db-55",
        "model": "ARC 1400 DICUT DB 55",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 55
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 269,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 271,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 266,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 263,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1400-dicut-db-65",
        "model": "ARC 1400 DICUT DB 65",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 65
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 259,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 261,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 256,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 253,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1400-dicut-db-85",
        "model": "ARC 1400 DICUT DB 85",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 85
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 239,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 241,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 236,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 233,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "erc-1100-dicut-35",
        "model": "ERC 1100 DICUT DB 35",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 35
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 285,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 288,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 287,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 285,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "erc-1400-dicut-35",
        "model": "ERC 1400 DICUT DB 35",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 35
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 285,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 288,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 287,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 285,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "erc-1100-dicut-45",
        "model": "ERC 1100 DICUT DB 45",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 45
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 275,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 278,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 277,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 275,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "erc-1400-dicut-45",
        "model": "ERC 1400 DICUT DB 45",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 45
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 275,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 278,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 277,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 275,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "grc-1100-dicut-30-700c",
        "model": "GRC 1100 DICUT DB 30 (700C)",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 30
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "grc-1400-dicut-30-700c",
        "model": "GRC 1400 DICUT DB 30 (700C)",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 30
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "grc-1100-dicut-50",
        "model": "GRC 1100 DICUT DB 50 (700C)",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 50
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 272,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 275,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 271,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "grc-1400-dicut-50",
        "model": "GRC 1400 DICUT DB 50 (700C)",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 50
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 270,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 275,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 274,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 270,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "grc-1100-dicut-30-650b",
        "model": "GRC 1100 DICUT DB 30 (650B)",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 30
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 271,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 271,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 273,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 271,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "grc-1400-dicut-30-650b",
        "model": "GRC 1400 DICUT DB 30 (650B)",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 30
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 271,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 271,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 270,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 270,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "prc-1400-spline-35",
        "model": "PRC 1400 SPLINE 35",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 35
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 20,
            "lacingPattern": "0X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 282,
                "spokeModel": "DT Aerolite T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 282,
                "spokeModel": "DT Aerolite T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 295,
                "spokeModel": "DT Aero Comp Straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 290,
                "spokeModel": "DT Aero Comp Straight-pull",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "xrc-1200-spline-29-30",
        "model": "XRC 1200 SPLINE 29″ 30",
        "lifecycleStatus": "current",
        "nippleModel": "DT ProLock Squorx ProHead Aluminum",
        "nippleLengthMm": 15,
        "rim": {
          "depthMm": 20
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 299,
                "spokeModel": "DT Revolite T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 300,
                "spokeModel": "DT Revolite T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 299,
                "spokeModel": "DT Revolite T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 299,
                "spokeModel": "DT Revolite T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "exc-1200-classic-27-5",
        "model": "EXC 1200 CLASSIC 27.5″ 30",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Flat Hexagonal Aluminum",
        "nippleLengthMm": 13,
        "rim": {
          "depthMm": 22
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 32,
            "lacingPattern": "3X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 32,
            "lacingPattern": "3X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "exc-1200-classic-29",
        "model": "EXC 1200 CLASSIC 29″ 30",
        "lifecycleStatus": "current",
        "nippleModel": "DT Pro Lock Flat Hexagonal Aluminum",
        "nippleLengthMm": 13,
        "rim": {
          "depthMm": 22
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 32,
            "lacingPattern": "3X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 296,
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 297,
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 32,
            "lacingPattern": "3X",
            "sides": [
              {
                "side": "left",
                "lengthMm": null,
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": null,
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1100-dicut-db-50-legacy",
        "model": "ARC 1100 DICUT DB 50",
        "lifecycleStatus": "legacy",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 15,
        "rim": {
          "depthMm": 50
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 272,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 273,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 273,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 270,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1100-dicut-db-62-legacy",
        "model": "ARC 1100 DICUT DB 62",
        "lifecycleStatus": "legacy",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 15,
        "rim": {
          "depthMm": 62
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 258,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 261,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 260,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 256,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1100-dicut-db-80-legacy",
        "model": "ARC 1100 DICUT DB 80",
        "lifecycleStatus": "legacy",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 15,
        "rim": {
          "depthMm": 80
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 238,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 238,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 238,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 233,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1400-dicut-db-50-legacy",
        "model": "ARC 1400 DICUT DB 50",
        "lifecycleStatus": "legacy",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 15,
        "rim": {
          "depthMm": 50
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 272,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 273,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 273,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 270,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "arc-1400-dicut-db-62-legacy",
        "model": "ARC 1400 DICUT DB 62",
        "lifecycleStatus": "legacy",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 15,
        "rim": {
          "depthMm": 62
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 258,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 261,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 260,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 256,
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "prc-1100-mon-chasseral-35-legacy",
        "model": "PRC 1100 DICUT Mon Chasseral 35 (Rim Brake)",
        "lifecycleStatus": "legacy",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 15,
        "rim": {
          "depthMm": 35
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 16,
            "lacingPattern": "0X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 283,
                "spokeModel": "DT Aerolite Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 283,
                "spokeModel": "DT Aerolite Straightpull",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 21,
            "lacingPattern": "2:1",
            "sides": [
              {
                "side": "left",
                "lengthMm": 286,
                "spokeModel": "DT Aerolite T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 284,
                "spokeModel": "DT Aero Comp Straightpull",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "prc-1100-mon-chasseral-24-db-legacy",
        "model": "PRC 1100 DICUT 24 DB Mon Chasseral",
        "lifecycleStatus": "legacy",
        "nippleModel": "DT Pro Lock Hidden Aluminum",
        "nippleLengthMm": 15,
        "rim": {
          "depthMm": 24
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 294,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 296,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 296,
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 294,
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "cr-1400-dicut-db-25-legacy",
        "model": "CR 1400 DICUT DB 25",
        "lifecycleStatus": "legacy",
        "nippleModel": "DT Pro Lock Squorx Pro Head Aluminum",
        "nippleLengthMm": 15,
        "rim": {
          "depthMm": 25
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 292,
                "spokeModel": "DT Aerolite Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 294,
                "spokeModel": "DT Aerolite Straightpull",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 294,
                "spokeModel": "DT Aerolite Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 290,
                "spokeModel": "DT Aero Comp Straightpull",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "pr-1600-dicut-21-rim-brake-legacy",
        "model": "PR 1600 DICUT 21（圈刹）",
        "lifecycleStatus": "legacy",
        "nippleModel": "DT Pro Lock Aluminum",
        "nippleLengthMm": 12,
        "rim": {
          "depthMm": 21
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 16,
            "lacingPattern": "0X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 282,
                "spokeModel": "DT Aero Comp Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 282,
                "spokeModel": "DT Aero Comp Straightpull",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 288,
                "spokeModel": "DT Aero Comp Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 292,
                "spokeModel": "DT Aero Comp Straightpull",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      }
    ]
  },
  {
    "brandSlug": "enve",
    "brandName": "ENVE",
    "publicationStatus": "published",
    "sourceCheckedAt": "2026-10-01",
    "wheelsets": [
      {
        "slug": "enve-ses-2-3-gen4",
        "model": "ENVE SES 2.3 (Gen 4)",
        "lifecycleStatus": "current",
        "nippleModel": "ENVE 7075-T6 inverted internal alloy nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 28,
          "depthFrontMm": 28,
          "depthRearMm": 32
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 300,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 300,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 296,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 296,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "enve-ses-3-4-gen4",
        "model": "ENVE SES 3.4 (Gen 4)",
        "lifecycleStatus": "current",
        "nippleModel": "ENVE molded inverted internal alloy nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 39,
          "depthFrontMm": 39,
          "depthRearMm": 43
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 287,
                "spokeModel": "Sapim CX-Ray TCS OH bladed 2.0-0.9 x 2.2-2.0 mm straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 287,
                "spokeModel": "Sapim CX-Ray TCS OH bladed 2.0-0.9 x 2.2-2.0 mm straight-pull",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 283,
                "spokeModel": "Sapim CX-Ray TCS OH bladed 2.0-0.9 x 2.2-2.0 mm straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 283,
                "spokeModel": "Sapim CX-Ray TCS OH bladed 2.0-0.9 x 2.2-2.0 mm straight-pull",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "enve-ses-4-5-gen4",
        "model": "ENVE SES 4.5 (Gen 4)",
        "lifecycleStatus": "current",
        "nippleModel": "ENVE inverted internal alloy nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 50,
          "depthFrontMm": 50,
          "depthRearMm": 56
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 276,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 276,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 271,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 271,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "enve-ses-6-7-gen4",
        "model": "ENVE SES 6.7 (Gen 4)",
        "lifecycleStatus": "current",
        "nippleModel": "ENVE inverted internal alloy nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 60,
          "depthFrontMm": 60,
          "depthRearMm": 67
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 267,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 267,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 261,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 261,
                "spokeModel": "Sapim CX-Ray TCS OH bladed straight-pull",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "enve-g23-gravel",
        "model": "ENVE G23 (700c Gravel)",
        "lifecycleStatus": "current",
        "nippleModel": "ENVE inverted internal alloy nipple 14G",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 25,
          "depthFrontMm": 25,
          "depthRearMm": 25
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 301,
                "spokeModel": "Sapim CX-Ray bladed straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 301,
                "spokeModel": "Sapim CX-Ray bladed straight-pull",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 301,
                "spokeModel": "Sapim CX-Ray bladed straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 301,
                "spokeModel": "Sapim CX-Ray bladed straight-pull",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "enve-foundation-45",
        "model": "ENVE 45 (Foundation Road)",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Double Square external alloy/brass nipple 14G",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 45,
          "depthFrontMm": 45,
          "depthRearMm": 45
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 274,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 276,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 276,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 270,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "enve-foundation-65",
        "model": "ENVE 65 (Foundation Road)",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Double Square external alloy/brass nipple 14G",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 65,
          "depthFrontMm": 65,
          "depthRearMm": 65
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 254,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 256,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 256,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 250,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "enve-foundation-ag25",
        "model": "ENVE AG25 (Foundation Gravel 700c)",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Double Square external brass nipple 14G",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 21,
          "depthFrontMm": 21,
          "depthRearMm": 21
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 290,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 294,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 292,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 288,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "enve-foundation-ag28",
        "model": "ENVE AG28 (Foundation Gravel 650b)",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Double Square external brass nipple 14G",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 21,
          "depthFrontMm": 21,
          "depthRearMm": 21
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 272,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 276,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 274,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 270,
                "spokeModel": "Sapim CX-Sprint bladed J-bend",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "enve-foundation-am30-29",
        "model": "ENVE AM30 (Foundation MTB 29\" Boost)",
        "lifecycleStatus": "current",
        "nippleModel": "Alpina Nylock external brass nipple 14G",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 20,
          "depthFrontMm": 20,
          "depthRearMm": 20
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 28,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 288,
                "spokeModel": "Sapim Race butted 2.0-1.8-2.0 mm J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 290,
                "spokeModel": "Sapim Race butted 2.0-1.8-2.0 mm J-bend",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 28,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 290,
                "spokeModel": "Sapim Race butted 2.0-1.8-2.0 mm J-bend",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 286,
                "spokeModel": "Sapim Race butted 2.0-1.8-2.0 mm J-bend",
                "headType": "j-bend"
              }
            ]
          }
        ]
      }
    ]
  },
  {
    "brandSlug": "shimano",
    "brandName": "Shimano",
    "publicationStatus": "published",
    "sourceCheckedAt": "2026-10-01",
    "wheelsets": [
      {
        "slug": "wh-r9270-c50-tl",
        "model": "DURA-ACE WH-R9270-C50-TL",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 50
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 267,
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 269,
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X / 2:1",
            "sides": [
              {
                "side": "left",
                "lengthMm": 251.5,
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 267.5,
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "wh-r9270-c36-tl",
        "model": "DURA-ACE WH-R9270-C36-TL",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 36
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 280.5,
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 282,
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X / 2:1",
            "sides": [
              {
                "side": "left",
                "lengthMm": 264,
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 279,
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "wh-r9270-c60-hr-tl",
        "model": "DURA-ACE WH-R9270-C60-HR-TL",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 60
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 248,
                "spokeModel": "Shimano DURA-ACE HR bladed 2.0-1.8-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 248,
                "spokeModel": "Shimano DURA-ACE HR bladed 2.0-1.8-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X / 2:1",
            "sides": [
              {
                "side": "left",
                "lengthMm": 241,
                "spokeModel": "Shimano DURA-ACE HR bladed 2.0-1.8-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 257.5,
                "spokeModel": "Shimano DURA-ACE HR bladed 2.0-1.8-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "wh-r8170-c50-tl",
        "model": "ULTEGRA WH-R8170-C50-TL",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 50
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 272,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 272,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X / 2:1",
            "sides": [
              {
                "side": "left",
                "lengthMm": 253,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 272,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "wh-r8170-c36-tl",
        "model": "ULTEGRA WH-R8170-C36-TL",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 36
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 286,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 286,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X / 2:1",
            "sides": [
              {
                "side": "left",
                "lengthMm": 265,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 284,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "wh-r8170-c60-tl",
        "model": "ULTEGRA WH-R8170-C60-TL",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 60
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 262,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 262,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X / 2:1",
            "sides": [
              {
                "side": "left",
                "lengthMm": 244,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 262,
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      },
      {
        "slug": "wh-rs710-c46-tl",
        "model": "105 WH-RS710-C46-TL",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 46
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 265.5,
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 267.5,
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 267.5,
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 265.5,
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "wh-rs710-c32-tl",
        "model": "105 WH-RS710-C32-TL",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 32
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 279,
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 281.5,
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 279,
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 281.5,
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "wh-rx880-tl",
        "model": "GRX WH-RX880-TL",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 32
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 281.5,
                "spokeModel": "Shimano GRX bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 283,
                "spokeModel": "Shimano GRX bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 280,
                "spokeModel": "Shimano GRX bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 279,
                "spokeModel": "Shimano GRX bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "wh-m8100-tl-29",
        "model": "DEORE XT WH-M8100-TL-29 (Boost)",
        "lifecycleStatus": "current",
        "nippleModel": "Shimano 14G aluminum nipple with spherical washer",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 18.8
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 28,
            "lacingPattern": "3X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 301.5,
                "spokeModel": "Shimano XT butted 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 301.5,
                "spokeModel": "Shimano XT butted 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 28,
            "lacingPattern": "3X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 301.5,
                "spokeModel": "Shimano XT butted 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "lengthMm": 298,
                "spokeModel": "Shimano XT butted 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      }
    ]
  },
  {
    "brandSlug": "zipp",
    "brandName": "ZIPP",
    "publicationStatus": "published",
    "sourceCheckedAt": "2026-10-01",
    "wheelsets": [
      {
        "slug": "zipp-303-firecrest-b1",
        "model": "ZIPP 303 Firecrest [B1 世代 · 2021–2026]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Secure Lock external black alloy nipple, 2.0 mm",
        "nippleLengthMm": 14,
        "rim": {
          "depthMm": 40
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 270,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 272,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 270,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 266,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-404-firecrest-b1",
        "model": "ZIPP 404 Firecrest [B1 世代 · 2021–2026]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Secure Lock external alloy nipple, 2.0 mm",
        "nippleLengthMm": 14,
        "rim": {
          "depthMm": 58
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 254,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 256,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 254,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 250,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-808-firecrest-b1",
        "model": "ZIPP 808 Firecrest [B1 世代 · 2022–2026]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Secure Lock external alloy nipple",
        "nippleLengthMm": 14,
        "rim": {
          "depthMm": 80
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 230,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 226,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 230,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 224,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-353-nsw-a1",
        "model": "ZIPP 353 NSW [A1 世代 · Cognition V2]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Secure Lock external alloy nipple",
        "nippleLengthMm": 14,
        "rim": {
          "depthMm": 45
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 264,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 266,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 266,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 260,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-454-nsw-b1-c1",
        "model": "ZIPP 454 NSW [B1/C1 世代 · Cognition V2]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Secure Lock external alloy nipple",
        "nippleLengthMm": 14,
        "rim": {
          "depthMm": 58
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 256,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 252,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 256,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 252,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-858-nsw-b1-d1",
        "model": "ZIPP 858 NSW [B1/D1 世代 · Cognition V2]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Secure Lock external alloy nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 85
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 232,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 226,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 20,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 230,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 232,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-303-xplr-sw-a1",
        "model": "ZIPP 303 XPLR SW [A1 世代 · 2024–2026]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Secure Lock external alloy nipple",
        "nippleLengthMm": 14,
        "rim": {
          "depthMm": 54
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 256,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 258,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 258,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 260,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-303-xplr-s-a1",
        "model": "ZIPP 303 XPLR S [A1 世代 · 2024–2026]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Secure Lock external brass/alloy nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 54
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 258,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 260,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 260,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 258,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-101-xplr-a1-700c",
        "model": "ZIPP 101 XPLR 700c [A1 世代 · MOTO]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim Secure Lock external alloy nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 15
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 28,
            "lacingPattern": "3X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 304,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 302,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 28,
            "lacingPattern": "3X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 304,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 302,
                "spokeModel": "Sapim CX-Sprint",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-303-s-a1",
        "model": "ZIPP 303 S [A1 世代]",
        "lifecycleStatus": "current",
        "nippleModel": "Sapim external brass/alloy nipple",
        "nippleLengthMm": null,
        "rim": {
          "depthMm": 45
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 266,
                "spokeModel": "Sapim CX54 / CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 268,
                "spokeModel": "Sapim CX54 / CX-Sprint",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 266,
                "spokeModel": "Sapim CX54 / CX-Sprint",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 264,
                "spokeModel": "Sapim CX54 / CX-Sprint",
                "headType": "j-bend"
              }
            ]
          }
        ]
      },
      {
        "slug": "zipp-303-firecrest-a1-legacy-77-177",
        "model": "ZIPP 303 Firecrest Carbon Clincher Disc [A1 世代 · MY16–MY19 · 77/177D]",
        "lifecycleStatus": "legacy",
        "nippleModel": "Sapim Secure Lock external nipple, 2.0 mm",
        "nippleLengthMm": 14,
        "rim": {
          "depthMm": 45
        },
        "wheels": [
          {
            "position": "front",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 272,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 274,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          },
          {
            "position": "rear",
            "spokeCount": 24,
            "lacingPattern": "2X",
            "sides": [
              {
                "side": "left",
                "lengthMm": 274,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              },
              {
                "side": "right",
                "lengthMm": 272,
                "spokeModel": "Sapim CX-Ray",
                "headType": "j-bend"
              }
            ]
          }
        ]
      }
    ]
  }
]
