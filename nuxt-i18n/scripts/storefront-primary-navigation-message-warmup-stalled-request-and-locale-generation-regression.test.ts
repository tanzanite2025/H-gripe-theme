import assert from 'node:assert/strict'
import { mock } from 'node:test'
import { setImmediate } from 'node:timers/promises'
import { warmStorefrontPrimaryNavigationMessagesWithBoundedNamespaceWait } from '../app/utils/warmStorefrontPrimaryNavigationMessagesWithBoundedNamespaceWait.js'

mock.timers.enable({ apis: ['setTimeout'] })
try {
  const warmedNamespaces: string[] = []
  await warmStorefrontPrimaryNavigationMessagesWithBoundedNamespaceWait(
    ['missing', 'healthy'],
    async namespace => {
      warmedNamespaces.push(namespace)
      if (namespace === 'missing') throw new Error('Download failed')
    },
    () => true,
  )
  assert.deepEqual(warmedNamespaces, ['missing', 'healthy'], 'A failed namespace must not stop later menus')

  const stalledWarmupStarts: string[] = []
  let rejectStalledDownload!: (error: Error) => void
  const stalledDownload = new Promise<void>((_resolve, reject) => { rejectStalledDownload = reject })
  const stalledWarmup = warmStorefrontPrimaryNavigationMessagesWithBoundedNamespaceWait(
    ['stalled', 'next-menu'],
    namespace => {
      stalledWarmupStarts.push(namespace)
      return namespace === 'stalled' ? stalledDownload : Promise.resolve()
    },
    () => true,
  )
  mock.timers.tick(1499)
  await setImmediate()
  assert.deepEqual(stalledWarmupStarts, ['stalled'], 'Normal downloads retain their sequential wait budget')
  mock.timers.tick(1)
  await stalledWarmup
  assert.deepEqual(stalledWarmupStarts, ['stalled', 'next-menu'], 'A hanging import cannot starve every later menu')
  rejectStalledDownload(new Error('Late download rejection'))
  await setImmediate()

  let headerMounted = true
  const unmountedWarmupStarts: string[] = []
  const unmountedWarmup = warmStorefrontPrimaryNavigationMessagesWithBoundedNamespaceWait(
    ['in-flight', 'must-not-start'],
    namespace => {
      unmountedWarmupStarts.push(namespace)
      return new Promise<void>(() => {})
    },
    () => headerMounted,
  )
  headerMounted = false
  mock.timers.tick(1500)
  await unmountedWarmup
  assert.deepEqual(unmountedWarmupStarts, ['in-flight'], 'Unmount stops starting further imports')

  let currentLocaleGeneration = 1
  let releaseOldLocaleDownloads!: () => void
  const oldLocaleDownload = new Promise<void>(resolve => { releaseOldLocaleDownloads = resolve })
  const localeWarmupStarts: string[] = []
  function startNavigationMessageWarmupForLocaleGeneration(generation: number) {
    return warmStorefrontPrimaryNavigationMessagesWithBoundedNamespaceWait(
      ['first', 'second'],
      namespace => {
        localeWarmupStarts.push(`${generation}:${namespace}`)
        return generation < 3 ? oldLocaleDownload : Promise.resolve()
      },
      () => generation === currentLocaleGeneration,
    )
  }
  const firstLocaleWarmup = startNavigationMessageWarmupForLocaleGeneration(1)
  currentLocaleGeneration = 2
  const secondLocaleWarmup = startNavigationMessageWarmupForLocaleGeneration(2)
  currentLocaleGeneration = 3
  await startNavigationMessageWarmupForLocaleGeneration(3)
  releaseOldLocaleDownloads()
  await Promise.all([firstLocaleWarmup, secondLocaleWarmup])
  assert.deepEqual(localeWarmupStarts, ['1:first', '2:first', '3:first', '3:second'],
    'Returning to the same locale must not restart superseded workers')
} finally {
  mock.timers.reset()
}

console.log('Primary navigation message warmup stalled-request and locale-generation regression checks passed.')
