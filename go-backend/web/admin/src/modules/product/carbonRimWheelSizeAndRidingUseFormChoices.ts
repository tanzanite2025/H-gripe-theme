export const carbonRimWheelSizeChoices = [
  { value: '700c', label: '700C', beadSeatDiameterMM: 622 },
  { value: '29-inch', label: '29 英寸', beadSeatDiameterMM: 622 },
  { value: '650b', label: '650B', beadSeatDiameterMM: 584 },
  { value: '27.5-inch', label: '27.5 英寸', beadSeatDiameterMM: 584 },
  { value: '650c', label: '650C', beadSeatDiameterMM: 571 },
  { value: '26-inch', label: '26 英寸', beadSeatDiameterMM: 559 },
  { value: '24-inch', label: '24 英寸', beadSeatDiameterMM: 507 },
  { value: '20-inch-406', label: '20 英寸（BSD 406）', beadSeatDiameterMM: 406 },
  { value: '20-inch-451', label: '20 英寸（BSD 451）', beadSeatDiameterMM: 451 },
] as const

export const carbonRimRidingUseChoices = [
  { value: 'road', label: '公路骑行' }, { value: 'gravel', label: 'Gravel 碎石骑行' },
  { value: 'cyclocross', label: '公路越野 CX' }, { value: 'cross-country', label: '山地越野 XC' },
  { value: 'trail', label: '山地林道 Trail' }, { value: 'enduro', label: '山地耐力 Enduro' },
  { value: 'downhill', label: '山地下坡 DH' }, { value: 'triathlon-time-trial', label: '铁人三项 / 计时赛' },
  { value: 'touring-bikepacking', label: '旅行 / Bikepacking' }, { value: 'urban-commuting', label: '城市通勤' },
] as const
