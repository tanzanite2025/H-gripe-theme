export type PublicBrandWheelsetPublicationStatus = 'draft' | 'published'
export type PublicBrandWheelsetLifecycleStatus = 'current' | 'legacy'
export type PublicBrandWheelPosition = 'front' | 'rear'
export type PublicBrandWheelSpokeSide = 'left' | 'right'
export type PublicBrandSpokeHeadType = 'straight-pull' | 'j-bend' | 'unknown'

export interface PublicBrandWheelSideSpokeSpec {
  side: PublicBrandWheelSpokeSide
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
  rim: { depthMm: number }
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp Straight-pull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Revolite T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Revolite T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Revolite",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite II T-head",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aerolite Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "DT Aero Comp Straightpull",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
    "brandSlug": "shimano",
    "brandName": "Shimano",
    "publicationStatus": "published",
    "sourceCheckedAt": "2026-10-01",
    "wheelsets": [
      {
        "slug": "wh-r9270-c50-tl",
        "model": "DURA-ACE WH-R9270-C50-TL",
        "lifecycleStatus": "current",
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
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano DURA-ACE bladed 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano DURA-ACE HR bladed 2.0-1.8-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano DURA-ACE HR bladed 2.0-1.8-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano ULTEGRA bladed 2.0-1.6-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano 105 bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano GRX bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano GRX bladed 2.0-1.6-2.0 mm",
                "headType": "j-bend"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano XT butted 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
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
                "spokeModel": "Shimano XT butted 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              },
              {
                "side": "right",
                "spokeModel": "Shimano XT butted 2.0-1.5-2.0 mm",
                "headType": "straight-pull"
              }
            ]
          }
        ]
      }
    ]
  }
]
