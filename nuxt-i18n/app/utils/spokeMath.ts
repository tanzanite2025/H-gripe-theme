/**
 * Frontend-only spoke length calculation
 * Formula: L = sqrt((ERD/2)^2 + (PCD/2)^2 + flange^2 - ERD * PCD/2 * cos(cross_angle))
 * where cross_angle = 4 * PI * crossing / spokeCount
 */
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
