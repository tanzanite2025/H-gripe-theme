import preflightApi, { type SiteQualityJob, type SiteQualityStrategy } from '@/api/preflight'

export const useSiteQualityJobs = () => {
  const enqueueInspection = async (
    url: string,
    strategy: SiteQualityStrategy,
    onUpdate?: (job: SiteQualityJob) => void,
  ): Promise<SiteQualityJob> => {
    const queued = await preflightApi.createSiteQualityJob(url, strategy)
    onUpdate?.(queued.job)
    return preflightApi.waitForSiteQualityJob(queued.job_id, { onUpdate })
  }

  const enqueueFindingRecheck = async (
    findingID: number,
    onUpdate?: (job: SiteQualityJob) => void,
  ): Promise<SiteQualityJob> => {
    const queued = await preflightApi.enqueueSiteQualityFindingRecheck(findingID)
    onUpdate?.(queued.job)
    return preflightApi.waitForSiteQualityJob(queued.job_id, { onUpdate })
  }

  return {
    enqueueInspection,
    enqueueFindingRecheck,
  }
}
