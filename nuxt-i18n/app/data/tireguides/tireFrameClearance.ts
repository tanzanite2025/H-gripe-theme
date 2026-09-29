export interface TireFrameClearanceRow {
  inch: string
  etrto: string
  tire: string
  maxWidthMm: number
  maxDiameterMm: number
  shoulderDiameterMm: number
}

export interface TireFrameClearanceTable {
  key: string
  title: string
  rows: readonly TireFrameClearanceRow[]
}

export const tireFrameClearanceTables: readonly TireFrameClearanceTable[] = [
  {
    key: '24-26-inch',
    title: '24 and 26 inch tire frame clearance',
    rows: [
      { inch: '24"', etrto: '60-507', tire: 'Crazy Bob', maxWidthMm: 61, maxDiameterMm: 631, shoulderDiameterMm: 570 },
      { inch: '26"', etrto: '60-559', tire: 'Big Apple', maxWidthMm: 58, maxDiameterMm: 683, shoulderDiameterMm: 625 },
      { inch: '26"', etrto: '60-559', tire: 'Big Ben', maxWidthMm: 60, maxDiameterMm: 689, shoulderDiameterMm: 628 },
      { inch: '26"', etrto: '60-559', tire: 'Crazy Bob', maxWidthMm: 64, maxDiameterMm: 685, shoulderDiameterMm: 629 },
      { inch: '26"', etrto: '60-559', tire: 'Dirty Dan', maxWidthMm: 65, maxDiameterMm: 694, shoulderDiameterMm: 623 },
      { inch: '26"', etrto: '60-559', tire: 'Fat Frank', maxWidthMm: 61, maxDiameterMm: 687, shoulderDiameterMm: 630 },
      { inch: '26"', etrto: '60-559', tire: 'Hans Dampf', maxWidthMm: 60, maxDiameterMm: 684, shoulderDiameterMm: 621 },
      { inch: '26"', etrto: '60-559', tire: 'Ice Spiker / Ice Spiker Pro', maxWidthMm: 60, maxDiameterMm: 686, shoulderDiameterMm: 615 },
      { inch: '26"', etrto: '60-559', tire: 'Magic Mary', maxWidthMm: 60, maxDiameterMm: 687, shoulderDiameterMm: 621 },
      { inch: '26"', etrto: '60-559', tire: 'Nobby Nic', maxWidthMm: 60, maxDiameterMm: 686, shoulderDiameterMm: 624 },
      { inch: '26"', etrto: '60-559', tire: 'Rock Razor', maxWidthMm: 60, maxDiameterMm: 683, shoulderDiameterMm: 620 },
      { inch: '26"', etrto: '60-559', tire: 'Rocket Ron', maxWidthMm: 60, maxDiameterMm: 687, shoulderDiameterMm: 628 },
      { inch: '26"', etrto: '60-559', tire: 'Space', maxWidthMm: 63, maxDiameterMm: 693, shoulderDiameterMm: 621 },
      { inch: '26"', etrto: '60-559', tire: 'Super Moto', maxWidthMm: 58, maxDiameterMm: 684, shoulderDiameterMm: 624 },
      { inch: '26"', etrto: '64-559', tire: 'Magic Mary', maxWidthMm: 67, maxDiameterMm: 701, shoulderDiameterMm: 636 },
    ],
  },
  {
    key: '27-29-inch',
    title: '27, 27.5, 28 and 29 inch tire frame clearance',
    rows: [
      { inch: '27.5"', etrto: '60-584', tire: 'Dirty Dan', maxWidthMm: 66, maxDiameterMm: 714, shoulderDiameterMm: 649 },
      { inch: '27.5"', etrto: '60-584', tire: 'Hans Dampf', maxWidthMm: 63, maxDiameterMm: 710, shoulderDiameterMm: 645 },
      { inch: '27.5"', etrto: '60-584', tire: 'Magic Mary', maxWidthMm: 62, maxDiameterMm: 713, shoulderDiameterMm: 641 },
      { inch: '27.5"', etrto: '60-584', tire: 'Nobby Nic', maxWidthMm: 62, maxDiameterMm: 712, shoulderDiameterMm: 648 },
      { inch: '27.5"', etrto: '60-584', tire: 'Rock Razor', maxWidthMm: 61, maxDiameterMm: 708, shoulderDiameterMm: 637 },
      { inch: '28"', etrto: '50-622', tire: 'Big Apple', maxWidthMm: 48, maxDiameterMm: 722, shoulderDiameterMm: 670 },
      { inch: '28"', etrto: '50-622', tire: 'Big Apple Plus', maxWidthMm: 51, maxDiameterMm: 729, shoulderDiameterMm: 677 },
      { inch: '28"', etrto: '55-622', tire: 'Big Apple', maxWidthMm: 55, maxDiameterMm: 741, shoulderDiameterMm: 688 },
      { inch: '28"', etrto: '55-622', tire: 'Big Ben', maxWidthMm: 57, maxDiameterMm: 744, shoulderDiameterMm: 688 },
      { inch: '28"', etrto: '55-622', tire: 'Marathon Almotion', maxWidthMm: 55, maxDiameterMm: 744, shoulderDiameterMm: 687 },
      { inch: '29"', etrto: '60-622', tire: 'Big Apple', maxWidthMm: 59, maxDiameterMm: 750, shoulderDiameterMm: 691 },
      { inch: '29"', etrto: '60-622', tire: 'Hans Dampf', maxWidthMm: 62, maxDiameterMm: 749, shoulderDiameterMm: 684 },
      { inch: '29"', etrto: '60-622', tire: 'Magic Mary', maxWidthMm: 61, maxDiameterMm: 751, shoulderDiameterMm: 678 },
      { inch: '29"', etrto: '60-622', tire: 'Nobby Nic', maxWidthMm: 59, maxDiameterMm: 751, shoulderDiameterMm: 687 },
      { inch: '29"', etrto: '60-622', tire: 'Racing Ralph', maxWidthMm: 59, maxDiameterMm: 747, shoulderDiameterMm: 686 },
      { inch: '29"', etrto: '60-622', tire: 'Super Moto', maxWidthMm: 59, maxDiameterMm: 750, shoulderDiameterMm: 691 },
    ],
  },
]
