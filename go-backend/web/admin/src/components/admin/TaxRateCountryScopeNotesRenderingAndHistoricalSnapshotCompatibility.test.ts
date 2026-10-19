import { createSSRApp } from 'vue'
import { renderToString } from 'vue/server-renderer'
import { describe, expect, it } from 'vitest'
import TaxRateCountryScopeNotes from './TaxRateCountryScopeNotes.vue'

describe('country tax rate scope notes', () => {
  it('distinguishes historical missing annotations from a collected empty source', async () => {
    for (const comments of [undefined, null]) {
      const html = await renderToString(createSSRApp(TaxRateCountryScopeNotes, { comments }))
      expect(html).toContain('此快照尚未采集适用范围附注')
      expect(html).not.toContain('本次采集的数据源未提供')
    }
    const html = await renderToString(createSSRApp(TaxRateCountryScopeNotes, { comments: {} }))
    expect(html).toContain('本次采集的数据源未提供适用范围附注')
    expect(html).not.toContain('尚未采集')
  })

  it('shows rate groups in numeric order, preserves continuation notes and escapes provider text', async () => {
    const html = await renderToString(createSSRApp(TaxRateCountryScopeNotes, {
      comments: { '13': ['Bicycle helmets only.\nHeading 6506.', 'Continuation <script>alert(1)</script>'], '8.5': ['Bicycle rental.'] },
    }))
    expect(html).toContain('适用范围与附注 · 3 条')
    expect(html.indexOf('8.5%')).toBeLessThan(html.indexOf('13%'))
    expect(html).toContain('Bicycle helmets only.\nHeading 6506.')
    expect(html.indexOf('Bicycle helmets')).toBeLessThan(html.indexOf('Continuation'))
    expect(html).toContain('&lt;script&gt;')
    expect(html).not.toContain('<script>')
    expect(html).toContain('同一税率下的附注不一定适用于每个类别')
  })
})
