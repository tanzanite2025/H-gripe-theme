export interface HubGeometry {
    leftFlange: number
    rightFlange: number
    leftFlangePcd: number
    rightFlangePcd: number
    spokeHoleDiameter?: number
}

export interface HubModel {
    id: string
    name: string
    // If a hub is distinct front/rear, it might only have one of these.
    // Or if it's a set sold together, it might have both.
    // We opt for a structure where a "Hub Product" can have spec for Front and/or Rear.
    front?: HubGeometry
    rear?: HubGeometry
}

export interface RimModel {
    id: string
    name: string
    erd: number
    weight?: number
}

export interface Brand<T> {
    id: string
    name: string
    items: T[]
}

export const RIM_DATABASE: Brand<RimModel>[] = [
    {
        id: 'dt_swiss',
        name: 'DT Swiss',
        items: [
            { id: 'rr411_db', name: 'RR 411 db', erd: 598 },
            { id: 'rr511_db', name: 'RR 511 db', erd: 581 },
            { id: 'rr421_db', name: 'RR 421 db', erd: 594 },
            { id: 'r460_db', name: 'R 460 db', erd: 596 },
            { id: 'gr531_db', name: 'GR 531 db', erd: 597 },
            { id: 'g540_db', name: 'G 540 db', erd: 592 },
        ],
    },
    {
        id: 'mavic',
        name: 'Mavic',
        items: [
            { id: 'open_pro_ust_disc', name: 'Open Pro UST Disc', erd: 589 },
            { id: 'open_pro_ust', name: 'Open Pro UST', erd: 589 },
            { id: 'a_1028', name: 'A 1028', erd: 614 },
        ],
    },
    {
        id: 'kinlin',
        name: 'Kinlin',
        items: [
            { id: 'xr26t', name: 'XR-26T', erd: 592 },
            { id: 'xr31t', name: 'XR-31T', erd: 580 },
        ],
    },
]

export const HUB_DATABASE: Brand<HubModel>[] = [
    {
        id: 'dt_swiss',
        name: 'DT Swiss',
        items: [
            {
                id: '180_road_db_cl',
                name: '180 Road db CL',
                front: { leftFlange: 22.5, rightFlange: 35.6, leftFlangePcd: 44, rightFlangePcd: 42 },
                rear: { leftFlange: 33, rightFlange: 20.2, leftFlangePcd: 46, rightFlangePcd: 46 },
            },
            {
                id: '240_road_db_cl',
                name: '240 EXP Road db CL',
                front: { leftFlange: 22.5, rightFlange: 35.6, leftFlangePcd: 44, rightFlangePcd: 42 },
                rear: { leftFlange: 33, rightFlange: 20.2, leftFlangePcd: 46, rightFlangePcd: 46 },
            },
            {
                id: '350_road_db_cl',
                name: '350 Road db CL',
                front: { leftFlange: 22.5, rightFlange: 35.6, leftFlangePcd: 44, rightFlangePcd: 42 },
                rear: { leftFlange: 33, rightFlange: 20.2, leftFlangePcd: 46, rightFlangePcd: 46 },
            },
            {
                id: '350_classic_db_is',
                name: '350 Classic db IS (6-bolt)',
                front: { leftFlange: 22.5, rightFlange: 35.6, leftFlangePcd: 58, rightFlangePcd: 45 },
                rear: { leftFlange: 35.5, rightFlange: 21.2, leftFlangePcd: 58, rightFlangePcd: 52 },
            },
        ],
    },
    {
        id: 'shimano',
        name: 'Shimano',
        items: [
            {
                id: 'hb_r7070',
                name: '105 HB-R7070',
                front: { leftFlange: 22, rightFlange: 35.6, leftFlangePcd: 44, rightFlangePcd: 44 },
                rear: { leftFlange: 36.5, rightFlange: 21.6, leftFlangePcd: 45, rightFlangePcd: 45 },
            },
        ],
    },
    {
        id: 'novatec',
        name: 'Novatec',
        items: [
            {
                id: 'd791sb_d792sb',
                name: 'D791SB / D792SB',
                front: { leftFlange: 27, rightFlange: 32, leftFlangePcd: 58, rightFlangePcd: 45 },
                rear: { leftFlange: 35, rightFlange: 21, leftFlangePcd: 58, rightFlangePcd: 49 },
            },
        ],
    },
]
