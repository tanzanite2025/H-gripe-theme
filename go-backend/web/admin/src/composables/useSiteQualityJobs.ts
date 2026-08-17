import preflightApi, { type SiteQualityJob, type SiteQualityStrategy } from '@/api/preflight'

export const useSiteQualityJobs = () => {
  const enqueueInspection = async (url: string, strategy: SiteQualityStrategy): Promise<SiteQualityJob> => {
    const queued = await preflightApi.createSiteQualityJob(url, strategy)
    return preflightApi.waitForSiteQualityJob(queued.job_id)
  }

  const enqueueFindingRecheck = async (findingID: number): Promise<SiteQualityJob> => {
    const queued = await preflightApi.enqueueSiteQualityFindingRecheck(findingID)
    return preflightApi.waitForSiteQualityJob(queued.job_id)
  }

  return {
    enqueueInspection,
    enqueueFindingRecheck,
  }
}
