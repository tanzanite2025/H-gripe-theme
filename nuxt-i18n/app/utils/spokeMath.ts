/**
 * Frontend-only spoke length calculation
 * Formula: L = sqrt((ERD/2)^2 + (PCD/2)^2 + flange^2 - ERD * PCD/2 * cos(cross_angle))
 * where cross_angle = 4 * PI * crossing / spokeCount
 */
export type SpokeSide = 'left' | 'right'

export interface SpokeTensionRatio {
    leftToRight: number
    rightToLeft: number
    lowerToHigher: number
    lowerSide: SpokeSide | 'balanced'
    leftBracingAngleDeg: number
    rightBracingAngleDeg: number
}

export function effectiveSpokeFlangeDistance(
    flangeDistance: number,
    rimOffsetMm: number,
    side: SpokeSide
): number {
    return side === 'left'
        ? flangeDistance + rimOffsetMm
        : flangeDistance - rimOffsetMm
}

export function computeSpokeLength(
    erd: number,
    flangePcd: number,
    flangeDistance: number,
    spokeCount: number,
    crossing: number,
    nippleType: 'standard' | 'hidden' = 'standard',
    nippleLength: number | null = null
): number {
    const erdRadius = erd / 2
    const pcdRadius = flangePcd / 2
    const crossAngle = (4 * Math.PI * crossing) / spokeCount

    // Standard spoke length formula based on triangle geometry
    const lengthSquared =
        erdRadius * erdRadius +
        pcdRadius * pcdRadius +
        flangeDistance * flangeDistance -
        2 * erdRadius * pcdRadius * Math.cos(crossAngle)

    let length = Math.sqrt(lengthSquared)

    // Hidden nipple correction: ADD length based on nipple depth
    // 9mm nipple → +6mm, 12mm nipple → +9mm (nipple length - 3)
    if (nippleType === 'hidden' && nippleLength) {
        const correction = nippleLength - 3
        length += correction
    }

    return Number(length.toFixed(1))
}

export function computeSpokeTensionRatio(
    leftBracingDistanceMm: number,
    rightBracingDistanceMm: number,
    leftLengthMm: number,
    rightLengthMm: number
): SpokeTensionRatio | null {
    if (
        leftBracingDistanceMm <= 0 ||
        rightBracingDistanceMm <= 0 ||
        leftLengthMm <= 0 ||
        rightLengthMm <= 0
    ) {
        return null
    }

    const leftSin = Math.min(1, leftBracingDistanceMm / leftLengthMm)
    const rightSin = Math.min(1, rightBracingDistanceMm / rightLengthMm)
    if (leftSin <= 0 || rightSin <= 0) return null

    const leftToRight = rightSin / leftSin
    const rightToLeft = leftSin / rightSin
    const lowerToHigher = Math.min(leftToRight, rightToLeft)

    let lowerSide: SpokeTensionRatio['lowerSide'] = 'balanced'
    if (leftToRight < 0.995) lowerSide = 'left'
    if (leftToRight > 1.005) lowerSide = 'right'

    return {
        leftToRight: roundSpokeRatio(leftToRight),
        rightToLeft: roundSpokeRatio(rightToLeft),
        lowerToHigher: roundSpokeRatio(lowerToHigher),
        lowerSide,
        leftBracingAngleDeg: roundSpokeRatio(Math.asin(leftSin) * 180 / Math.PI),
        rightBracingAngleDeg: roundSpokeRatio(Math.asin(rightSin) * 180 / Math.PI),
    }
}

function roundSpokeRatio(value: number) {
    return Number(value.toFixed(4))
}
