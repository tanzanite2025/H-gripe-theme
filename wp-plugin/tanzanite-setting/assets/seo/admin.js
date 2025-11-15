(function () {
    const { createElement: h, Fragment, useState, useEffect, useMemo } = wp.element
  const { __ } = wp.i18n
  const {
    Button,
    TextControl,
    TextareaControl,
    PanelBody,
    Panel,
    Spinner,
    Notice,
    Card,
    CardBody,
    CardHeader,
    CheckboxControl,
    SelectControl
  } = wp.components
  const apiFetch = wp.apiFetch

  // Simple template renderer for MVP
  const renderTemplate = (tpl, ctx) => {
    const map = ctx || {}
    const source = String(tpl || '')
    return source.replace(/\{(\w+)\}/g, (_, k) => (map[k] != null ? String(map[k]) : ''))
  }

  // Product Schema Field Mapping panel
  const ProductSchemaMappingPanel = () => {
    const initial = (MyThemeSEO.settings && MyThemeSEO.settings.schema && MyThemeSEO.settings.schema.product) || {}
    const languages = Array.isArray(MyThemeSEO.languages) ? MyThemeSEO.languages : []
    const [mappings, setMappings] = useState(Array.isArray(initial.mappings) ? initial.mappings : [])
    const [saving, setSaving] = useState(false)
    const [note, setNote] = useState(null)

    const fieldDefs = [
      { key: 'name', label: __('Title (name)', 'mytheme-seo'), type: 'text' },
      { key: 'description', label: __('Description', 'mytheme-seo'), type: 'text' },
      { key: 'image', label: __('Image URL', 'mytheme-seo'), type: 'image' },
      { key: 'brand', label: __('Brand', 'mytheme-seo'), type: 'text' },
      { key: 'sku', label: __('SKU', 'mytheme-seo'), type: 'text' },
      { key: 'gtin', label: __('GTIN', 'mytheme-seo'), type: 'text' },
      { key: 'mpn', label: __('MPN', 'mytheme-seo'), type: 'text' },
    ]

    const getRow = (field) => mappings.find((m) => m.field === field) || { field, source: { type: 'none' }, transforms: [] }
    const setRow = (field, updater) => {
      setMappings((prev) => {
        const next = [...prev]
        const i = next.findIndex((m) => m.field === field)
        const curr = i >= 0 ? next[i] : { field, source: { type: 'none' }, transforms: [] }
        const updated = updater(curr)
        if (i >= 0) next[i] = updated
        else next.push(updated)
        return next
      })
    }

    const toggleTransform = (field, t) => setRow(field, (r) => {
      const on = Array.isArray(r.transforms) ? r.transforms : []
      const has = on.includes(t)
      return { ...r, transforms: has ? on.filter((x) => x !== t) : [...on, t] }
    })

    const save = async () => {
      setSaving(true)
      setNote(null)
      try {
        await apiFetch({ path: '/mytheme/v1/seo/settings', method: 'POST', data: { settings: { schema: { product: { mappings } } } } })
        setNote({ status: 'success', message: __('已保存字段映射。', 'mytheme-seo') })
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('保存失败。', 'mytheme-seo') })
      } finally {
        setSaving(false)
      }
    }

    const fillRecommended = () => {
      const rec = [
        { field: 'name', source: { type: 'wc', wc: 'name' }, transforms: ['trim'] },
        { field: 'description', source: { type: 'core', key: 'post_excerpt' }, transforms: ['strip_tags','trim'] },
        { field: 'image', source: { type: 'wc', wc: 'image_id' }, transforms: ['id_to_url'] },
        { field: 'sku', source: { type: 'wc', wc: 'sku' }, transforms: ['trim'] },
        { field: 'gtin', source: { type: 'meta', key: '_gtin' }, transforms: ['trim'] },
        { field: 'mpn', source: { type: 'meta', key: '_mpn' }, transforms: ['trim'] },
      ]
      setMappings(rec)
      setNote({ status: 'success', message: __('已填充推荐映射（可修改后保存）。', 'mytheme-seo') })
    }

    // Preview with source labels
    const [pvId, setPvId] = useState('')
    const [pvSlug, setPvSlug] = useState('')
    const [pvLocale, setPvLocale] = useState(languages[0] || '')
    const [pvData, setPvData] = useState(null)
    const [pvOver, setPvOver] = useState(null)
    const [loading, setLoading] = useState(false)

    const doPreview = async () => {
      setLoading(true)
      setPvData(null)
      setPvOver(null)
      try {
        let schemaPath = ''
        if (pvId && String(pvId).trim() !== '') schemaPath = `/mytheme/v1/seo/schema/product/${encodeURIComponent(pvId)}${pvLocale ? `?locale=${encodeURIComponent(pvLocale)}` : ''}`
        else if (pvSlug && String(pvSlug).trim() !== '') schemaPath = `/mytheme/v1/seo/schema/product/by-slug/${encodeURIComponent(pvSlug)}${pvLocale ? `?locale=${encodeURIComponent(pvLocale)}` : ''}`
        if (!schemaPath) { setLoading(false); return }
        const res = await apiFetch({ path: schemaPath })
        setPvData(res && res.schema ? res.schema : res)
        if (pvId) {
          try { setPvOver(await apiFetch({ path: `/mytheme/v1/seo/product/overrides/${encodeURIComponent(pvId)}` })) } catch {}
        }
      } catch (e) {
        console.error(e)
      } finally {
        setLoading(false)
      }
    }

    const guessLabel = (field) => {
      // approximate label: override for brand if override exists; else mapping if mapping configured; else default
      if (field === 'brand' && pvOver && pvOver.data && pvOver.data.brand && String(pvOver.data.brand).trim() !== '') return __('覆盖', 'mytheme-seo')
      const hasMap = !!mappings.find((m) => m.field === field && m.source && m.source.type && m.source.type !== 'none')
      return hasMap ? __('映射', 'mytheme-seo') : __('默认', 'mytheme-seo')
    }

    const transformOptions = [
      { key: 'strip_tags', label: 'strip_tags' },
      { key: 'trim', label: 'trim' },
      { key: 'to_number', label: 'to_number' },
      { key: 'id_to_url', label: 'id_to_url' },
      { key: 'first_gallery_to_url', label: 'first_gallery_to_url' },
    ]

    const renderRow = (def) => {
      const row = getRow(def.key)
      const src = row.source || { type: 'none' }
      return h('div', { key: def.key, className: 'mytheme-seo-map-row', style: { padding: '10px 12px', border: '1px solid #eee', borderRadius: '6px', marginBottom: '8px' } },
        h('div', { style: { display: 'grid', gridTemplateColumns: '180px 1fr' } },
          h('div', null, h('strong', null, def.label)),
          h('div', null,
            h(SelectControl, {
              label: __('来源类型', 'mytheme-seo'),
              value: src.type || 'none',
              onChange: (v) => setRow(def.key, (r) => ({ ...r, source: { type: v } })) ,
              options: [
                { label: '—', value: 'none' },
                { label: 'Core', value: 'core' },
                { label: 'Meta', value: 'meta' },
                { label: 'WooCommerce', value: 'wc' },
                { label: 'Constant', value: 'const' },
              ]
            }),
            (src.type === 'core') && h(SelectControl, {
              label: __('核心字段', 'mytheme-seo'),
              value: src.key || 'post_title',
              onChange: (v) => setRow(def.key, (r) => ({ ...r, source: { ...r.source, key: v } })),
              options: [
                { label: 'post_title', value: 'post_title' },
                { label: 'post_excerpt', value: 'post_excerpt' },
                { label: 'post_content', value: 'post_content' },
                { label: 'post_author_display_name', value: 'post_author_display_name' },
              ]
            }),
            (src.type === 'meta') && h(TextControl, { label: __('Meta Key', 'mytheme-seo'), value: src.key || '', onChange: (v) => setRow(def.key, (r) => ({ ...r, source: { ...r.source, key: v } })) }),
            (src.type === 'wc') && h(SelectControl, {
              label: __('WooCommerce 字段', 'mytheme-seo'),
              value: src.wc || 'name',
              onChange: (v) => setRow(def.key, (r) => ({ ...r, source: { ...r.source, wc: v } })),
              options: [
                { label: 'name', value: 'name' },
                { label: 'sku', value: 'sku' },
                { label: 'price', value: 'price' },
                { label: 'regular_price', value: 'regular_price' },
                { label: 'sale_price', value: 'sale_price' },
                { label: 'stock_status', value: 'stock_status' },
                { label: 'image_id', value: 'image_id' },
                { label: 'gallery_ids', value: 'gallery_ids' },
              ]
            }),
            (src.type === 'const') && (
              h('div', null,
                languages.length > 0 && h('div', { style: { display: 'grid', gridTemplateColumns: '120px 1fr', gap: '8px', alignItems: 'center' } },
                  languages.map((loc) => h('div', { key: loc, style: { display: 'contents' } },
                    h('label', null, String(loc).toUpperCase()),
                    h(TextControl, { value: (src.i18n && src.i18n[loc]) || '', onChange: (v) => setRow(def.key, (r) => ({ ...r, source: { ...r.source, i18n: { ...(r.source?.i18n || {}), [loc]: v } } })) })
                  ))
                ),
                h(TextControl, { label: __('常量（回退）', 'mytheme-seo'), value: src.value || '', onChange: (v) => setRow(def.key, (r) => ({ ...r, source: { ...r.source, value: v } })) })
              )
            ),
            h('div', { style: { marginTop: '4px' } },
              h('label', null, __('转换', 'mytheme-seo')),
              h('div', { style: { display: 'flex', gap: '12px', flexWrap: 'wrap' } },
                transformOptions.map((op) => h(CheckboxControl, {
                  key: op.key,
                  label: op.label,
                  checked: Array.isArray(row.transforms) ? row.transforms.includes(op.key) : false,
                  onChange: () => toggleTransform(def.key, op.key)
                }))
              )
            )
          )
        )
      )
    }

    return h(Card, { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('字段映射（Product Schema）', 'mytheme-seo')),
      h(CardBody, null,
        h('div', null, fieldDefs.map(renderRow)),
        h('div', { style: { marginTop: '8px', display: 'flex', gap: '8px' } },
          h(Button, { variant: 'secondary', onClick: fillRecommended }, __('填充推荐映射', 'mytheme-seo')),
          h(Button, { variant: 'primary', onClick: save, disabled: saving }, saving ? __('保存中…', 'mytheme-seo') : __('保存字段映射', 'mytheme-seo'))
        ),
        note && h(Notice, { status: note.status, onRemove: () => setNote(null), isDismissible: true }, note.message),
        h('hr'),
        h('h4', null, __('预览（含来源标注）', 'mytheme-seo')),
        h('div', { style: { display: 'grid', gridTemplateColumns: '1fr 1fr 160px auto', gap: '8px', alignItems: 'end' } },
          h(TextControl, { label: 'Product ID', value: pvId, onChange: setPvId }),
          h(TextControl, { label: 'Product slug', value: pvSlug, onChange: setPvSlug }),
          h(SelectControl, { label: __('语言', 'mytheme-seo'), value: pvLocale, onChange: setPvLocale, options: languages.map((loc) => ({ label: String(loc).toUpperCase(), value: loc })) }),
          h(Button, { variant: 'secondary', onClick: doPreview, disabled: loading }, loading ? __('加载中…', 'mytheme-seo') : __('预览', 'mytheme-seo')),
        ),
        pvData && h('div', { style: { marginTop: '8px' } },
          h('table', { className: 'widefat striped' },
            h('thead', null, h('tr', null, h('th', null, __('字段', 'mytheme-seo')), h('th', null, __('值', 'mytheme-seo')), h('th', null, __('来源', 'mytheme-seo')))),
            h('tbody', null,
              fieldDefs.map((fd) => h('tr', { key: fd.key },
                h('td', null, fd.label),
                h('td', null, typeof pvData[fd.key] === 'object' ? JSON.stringify(pvData[fd.key]) : String(pvData[fd.key] ?? '')),
                h('td', null, guessLabel(fd.key))
              ))
            )
          )
        )
      )
    )
  }

  // Product Schema settings panel
  const ProductSchemaPanel = () => {
    const initial = (MyThemeSEO.settings && MyThemeSEO.settings.schema && MyThemeSEO.settings.schema.product) || {}
    const [cfg, setCfg] = useState({
      enabled: initial.enabled !== false,
      brand: initial.brand || '',
      priceSource: initial.priceSource || 'sale_or_regular',
      brand_i18n: initial.brand_i18n || {},
    })
    const [saving, setSaving] = useState(false)
    const [note, setNote] = useState(null)
    const [previewId, setPreviewId] = useState('')
    const [previewSlug, setPreviewSlug] = useState('')
    const languages = Array.isArray(MyThemeSEO.languages) ? MyThemeSEO.languages : []
    const [pvLocale, setPvLocale] = useState(languages[0] || '')
    const [pvLoading, setPvLoading] = useState(false)
    const [pvData, setPvData] = useState(null)

    // Per-product overrides editor
    const [ovId, setOvId] = useState('')
    const [ovBrand, setOvBrand] = useState('')
    const [ovPriceSource, setOvPriceSource] = useState('sale_or_regular')
    const [ovNote, setOvNote] = useState(null)

    const save = async () => {
      setSaving(true)
      setNote(null)
      try {
        await apiFetch({ path: '/mytheme/v1/seo/settings', method: 'POST', data: { settings: { schema: { product: cfg } } } })
        setNote({ status: 'success', message: __('已保存 Product Schema 设置。', 'mytheme-seo') })
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('保存失败。', 'mytheme-seo') })
      } finally {
        setSaving(false)
      }
    }

    const preview = async () => {
      setPvLoading(true)
      setPvData(null)
      try {
        let path = ''
        if (previewId && String(previewId).trim() !== '') {
          path = `/mytheme/v1/seo/schema/product/${encodeURIComponent(previewId)}${pvLocale ? `?locale=${encodeURIComponent(pvLocale)}` : ''}`
        } else if (previewSlug && String(previewSlug).trim() !== '') {
          path = `/mytheme/v1/seo/schema/product/by-slug/${encodeURIComponent(previewSlug)}${pvLocale ? `?locale=${encodeURIComponent(pvLocale)}` : ''}`
        } else {
          setPvLoading(false)
          setNote({ status: 'error', message: __('请填写 ID 或 slug 进行预览。', 'mytheme-seo') })
          return
        }
        const res = await apiFetch({ path })
        setPvData(res && res.schema ? res.schema : (res || null))
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('预览失败。', 'mytheme-seo') })
      } finally {
        setPvLoading(false)
      }
    }

    return h(Card, { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('Product Schema（结构化数据）', 'mytheme-seo')),
      h(CardBody, null,
        h(CheckboxControl, { label: __('启用 Product JSON-LD', 'mytheme-seo'), checked: !!cfg.enabled, onChange: (v) => setCfg((c) => ({ ...c, enabled: !!v })) }),
        h(TextControl, { label: __('默认品牌（可选）', 'mytheme-seo'), value: cfg.brand, onChange: (v) => setCfg((c) => ({ ...c, brand: v })) }),
        h(SelectControl, { label: __('价格来源', 'mytheme-seo'), value: cfg.priceSource, onChange: (v) => setCfg((c) => ({ ...c, priceSource: v })), options: [
          { label: __('促销价优先（有则用促销价，否则用常规价）', 'mytheme-seo'), value: 'sale_or_regular' },
          { label: __('仅用常规价', 'mytheme-seo'), value: 'regular_only' }
        ] }),
        // Multilingual brand editor
        languages.length > 0 && h('div', { style: { marginTop: '8px' } },
          h('h4', null, __('多语言品牌', 'mytheme-seo')),
          h('div', null,
            languages.map((loc) => h('div', { key: loc, style: { display: 'grid', gridTemplateColumns: '120px 1fr', gap: '8px', alignItems: 'center', marginBottom: '6px' } },
              h('label', null, String(loc).toUpperCase()),
              h(TextControl, { value: (cfg.brand_i18n && cfg.brand_i18n[loc]) || '', onChange: (v) => setCfg((c) => ({ ...c, brand_i18n: { ...(c.brand_i18n || {}), [loc]: v } })) })
            ))
          )
        ),
        h(Button, { variant: 'primary', onClick: save, disabled: saving, style: { marginTop: '8px' } }, saving ? __('保存中…', 'mytheme-seo') : __('保存设置', 'mytheme-seo')),
        note && h(Notice, { status: note.status, onRemove: () => setNote(null), isDismissible: true }, note.message),
        h('hr'),
        h('h4', null, __('预览 JSON-LD', 'mytheme-seo')),
        h('div', { style: { display: 'grid', gridTemplateColumns: '1fr 1fr 160px auto', gap: '8px', alignItems: 'end' } },
          h(TextControl, { label: 'Product ID', value: previewId, onChange: setPreviewId }),
          h(TextControl, { label: 'Product slug', value: previewSlug, onChange: setPreviewSlug }),
          h(SelectControl, { label: __('语言', 'mytheme-seo'), value: pvLocale, onChange: setPvLocale, options: languages.map((loc) => ({ label: String(loc).toUpperCase(), value: loc })) }),
          h(Button, { variant: 'secondary', onClick: preview, disabled: pvLoading }, pvLoading ? __('加载中…', 'mytheme-seo') : __('预览', 'mytheme-seo')),
        ),
        pvData && h('pre', { style: { marginTop: '8px', maxHeight: '300px', overflow: 'auto' } }, JSON.stringify(pvData, null, 2)),
        h('hr'),
        h('h4', null, __('单商品覆盖（Brand/价格来源）', 'mytheme-seo')),
        h('div', { style: { display: 'grid', gridTemplateColumns: '1fr 1fr 1fr auto', gap: '8px', alignItems: 'end' } },
          h(TextControl, { label: 'Product ID', value: ovId, onChange: setOvId }),
          h(TextControl, { label: __('覆盖品牌（可选）', 'mytheme-seo'), value: ovBrand, onChange: setOvBrand }),
          h(SelectControl, { label: __('覆盖价格来源', 'mytheme-seo'), value: ovPriceSource, onChange: setOvPriceSource, options: [
            { label: __('促销价优先', 'mytheme-seo'), value: 'sale_or_regular' },
            { label: __('仅用常规价', 'mytheme-seo'), value: 'regular_only' }
          ] }),
          h(Button, { variant: 'secondary', onClick: async () => {
            try {
              const idNum = parseInt(ovId, 10)
              if (!idNum) { setOvNote({ status: 'error', message: __('请输入有效 Product ID', 'mytheme-seo') }); return }
              const res = await apiFetch({ path: `/mytheme/v1/seo/product/overrides/${idNum}`, method: 'POST', data: { overrides: { brand: ovBrand, priceSource: ovPriceSource } } })
              setOvNote({ status: 'success', message: __('已保存覆盖。', 'mytheme-seo') })
            } catch (e) {
              console.error(e)
              setOvNote({ status: 'error', message: __('保存覆盖失败。', 'mytheme-seo') })
            }
          } }, __('保存覆盖', 'mytheme-seo')),
        ),
        h('div', { style: { marginTop: '6px', display: 'flex', gap: '8px' } },
          h(Button, { variant: 'secondary', onClick: async () => {
            try {
              const idNum = parseInt(ovId, 10)
              if (!idNum) { setOvNote({ status: 'error', message: __('请输入有效 Product ID', 'mytheme-seo') }); return }
              const res = await apiFetch({ path: `/mytheme/v1/seo/product/overrides/${idNum}` })
              const data = res && res.data ? res.data : {}
              setOvBrand(data.brand || '')
              setOvPriceSource(data.priceSource || 'sale_or_regular')
              setOvNote({ status: 'success', message: __('已加载覆盖。', 'mytheme-seo') })
            } catch (e) {
              console.error(e)
              setOvNote({ status: 'error', message: __('加载覆盖失败。', 'mytheme-seo') })
            }
          } }, __('读取覆盖', 'mytheme-seo')),
          ovNote && h(Notice, { status: ovNote.status, onRemove: () => setOvNote(null), isDismissible: true }, ovNote.message)
        )
      )
    )
  }

  // Extract brand/price using schema.org friendly paths
  const extractBrandPrice = (jsonld) => {
    const j = (jsonld && typeof jsonld === 'object') ? jsonld : {}
    let brand = ''
    let price = ''
    // brand can be string or object { name }
    if (typeof j.brand === 'string') brand = j.brand
    else if (j.brand && typeof j.brand.name === 'string') brand = j.brand.name

    // price prefer offers.price, support array offers[0]
    if (j.offers) {
      const offers = Array.isArray(j.offers) ? j.offers[0] : j.offers
      if (offers && (typeof offers.price === 'string' || typeof offers.price === 'number')) {
        price = String(offers.price)
      }
    }
    // fallback to direct price
    if (!price && (typeof j.price === 'string' || typeof j.price === 'number')) price = String(j.price)

    return { brand, price }
  }


  // Wrapper controls to opt into next styles globally where used
  // 不再强制 40px 高度，交由自定义 CSS 控制为 30px，以消除点击后出现更高的半透明区域与对齐问题
  const NextSelect = (props) => h(SelectControl, { __nextHasNoMarginBottom: true, ...props })
  const NextText = (props) => h(TextControl, { __nextHasNoMarginBottom: true, ...props })
  const NextTextarea = (props) => h(TextareaControl, { __nextHasNoMarginBottom: true, ...props })

  // Sitemaps settings panel (MVP)
  const SitemapsPanel = () => {
    const initial = MyThemeSEO.settings?.sitemaps || {}
    const [cfg, setCfg] = useState({
      enabled: !!initial.enabled,
      splitByLocale: initial.splitByLocale ?? true,
      includeImages: !!initial.includeImages,
      includeVideos: !!initial.includeVideos,
      pingOnRebuild: initial.pingOnRebuild ?? true,
      types: {
        pages: initial.types?.pages ?? true,
        posts: initial.types?.posts ?? true,
        products: initial.types?.products ?? true,
        categories: initial.types?.categories ?? true,
        product_categories: initial.types?.product_categories ?? true
      },
      externalUrls: Array.isArray(initial.externalUrls) ? initial.externalUrls : []
    })
    const [note, setNote] = useState(null)
    const [busy, setBusy] = useState(false)
    const [result, setResult] = useState(null)
    const [extStatus, setExtStatus] = useState({}) // { url: 'OK'|'FAIL'|'...' }

    const save = async () => {
      setBusy(true)
      setNote(null)
      try {
        await apiFetch({ path: '/mytheme/v1/seo/settings', method: 'POST', data: { settings: { sitemaps: cfg } } })
        setNote({ status: 'success', message: __('Sitemaps 设置已保存。', 'mytheme-seo') })
        // 保存后自动检测一次
        await checkExternalUrls()
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('保存失败。', 'mytheme-seo') })
      } finally {
        setBusy(false)
      }
    }

    const rebuild = async () => {
      setBusy(true)
      setResult(null)
      setNote(null)
      try {
        const res = await apiFetch({ path: '/mytheme/v1/seo/sitemaps/rebuild', method: 'POST', data: { full: true } })
        setResult(res)
        setNote({ status: 'success', message: __('已触发重建。', 'mytheme-seo') })
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('重建失败。', 'mytheme-seo') })
      } finally {
        setBusy(false)
      }
    }

    const ping = async () => {
      setBusy(true)
      setNote(null)
      try {
        const res = await apiFetch({ path: '/mytheme/v1/seo/sitemaps/ping', method: 'POST', data: { engines: ['google','bing'] } })
        setResult(res)
        setNote({ status: 'success', message: __('已提交 Ping 请求。', 'mytheme-seo') })
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('Ping 失败。', 'mytheme-seo') })
      } finally {
        setBusy(false)
      }
    }

    const setType = (k, v) => setCfg((c) => ({ ...c, types: { ...(c.types || {}), [k]: v } }))

    const siteOrigin = (typeof window !== 'undefined' ? window.location.origin : '')
    const sitemapIndexUrl = siteOrigin ? `${siteOrigin}/sitemap_index.xml` : '/sitemap_index.xml'

    return h(Card, { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('Sitemaps', 'mytheme-seo')),
      h(CardBody, null,
        h(CheckboxControl, { label: __('启用 Sitemaps', 'mytheme-seo'), checked: cfg.enabled, onChange: (v) => setCfg((c) => ({ ...c, enabled: v })) }),
        h(CheckboxControl, { label: __('按语言拆分', 'mytheme-seo'), checked: cfg.splitByLocale, onChange: (v) => setCfg((c) => ({ ...c, splitByLocale: v })) }),
        h(CheckboxControl, { label: __('包含图片 sitemap', 'mytheme-seo'), checked: cfg.includeImages, onChange: (v) => setCfg((c) => ({ ...c, includeImages: v })) }),
        h(CheckboxControl, { label: __('包含视频 sitemap', 'mytheme-seo'), checked: cfg.includeVideos, onChange: (v) => setCfg((c) => ({ ...c, includeVideos: v })) }),
        h(CheckboxControl, { label: __('重建后自动 Ping', 'mytheme-seo'), checked: cfg.pingOnRebuild, onChange: (v) => setCfg((c) => ({ ...c, pingOnRebuild: v })) }),
        h('hr'),
        h('h4', null, __('包含的类型', 'mytheme-seo')),
        h(CheckboxControl, { label: 'Pages', checked: cfg.types.pages, onChange: (v) => setType('pages', v) }),
        h(CheckboxControl, { label: 'Posts', checked: cfg.types.posts, onChange: (v) => setType('posts', v) }),
        h(CheckboxControl, { label: 'Products', checked: cfg.types.products, onChange: (v) => setType('products', v) }),
        h(CheckboxControl, { label: 'Categories', checked: cfg.types.categories, onChange: (v) => setType('categories', v) }),
        h(CheckboxControl, { label: 'Product Categories', checked: cfg.types.product_categories, onChange: (v) => setType('product_categories', v) }),
        h('div', { style: { marginTop: '8px' } },
          h(Button, { variant: 'primary', onClick: save, disabled: busy, style: { marginRight: '8px' } }, busy ? __('处理中…', 'mytheme-seo') : __('保存设置', 'mytheme-seo')),
          h(Button, { variant: 'secondary', onClick: rebuild, disabled: busy, style: { marginRight: '8px' } }, __('立即重建', 'mytheme-seo')),
          h(Button, { variant: 'secondary', onClick: ping, disabled: busy }, __('立即 Ping', 'mytheme-seo'))
        ),
        h('hr'),
        h('h4', null, __('外部子 sitemap 列表', 'mytheme-seo')),
        h('p', { style: { color: '#6b7280', marginTop: 0 } }, __('每行一个绝对 URL（http/https），保存后会纳入 /sitemap_index.xml。', 'mytheme-seo')),
        h('textarea', {
          value: (cfg.externalUrls || []).join('\n'),
          onChange: (e) => {
            const lines = String(e.target.value || '').split(/\r?\n/)
            setCfg((c) => ({ ...c, externalUrls: lines }))
          },
          style: { width: '100%', minHeight: '100px', borderRadius: '8px', border: '1px solid #e5e7eb', padding: '8px' }
        }),
        h('div', { style: { marginTop: '6px', color: '#374151' } },
          (cfg.externalUrls || []).filter(Boolean).map((u, i) =>
            h('div', { key: i, style: { display: 'flex', alignItems: 'center', gap: '8px' } },
              h('a', { href: u, target: '_blank', rel: 'noopener noreferrer' }, u),
              h('span', { 'data-url': u, className: 'ext-sitemap-status', style: { fontSize: '12px', color: (extStatus[u]==='OK'?'#059669': extStatus[u]==='FAIL'?'#d14343':'#6b7280') } }, extStatus[u] || '…')
            )
          )
        ),
        h('div', { style: { marginTop: '6px' } },
          h(Button, { variant: 'secondary', onClick: checkExternalUrls, disabled: busy }, __('检测可用性', 'mytheme-seo'))
        ),
        // Show where the sitemap lives (URL). Core/most plugins generate virtual sitemap; link is the canonical location.
        h('p', { style: { marginTop: '8px', color: '#374151' } },
          h('strong', null, __('索引地址：', 'mytheme-seo')),
          ' ', h('a', { href: sitemapIndexUrl, target: '_blank', rel: 'noopener noreferrer' }, sitemapIndexUrl),
          ' ', h('span', { style: { color: '#6b7280' } }, __('（多数情况下为虚拟生成，不落地文件）', 'mytheme-seo'))
        ),
        result && h('div', { style: { marginTop: '8px' } },
          h('pre', null, JSON.stringify(result, null, 2))
        ),
        note && h(Notice, { status: note.status, onRemove: () => setNote(null), isDismissible: true }, note.message)
      )
    )
  }

  const pushIndexNow = async (urls) => {
    try {
      const list = Array.isArray(urls) ? urls.filter(Boolean) : []
      if (!list.length) return
      await apiFetch({ path: '/mytheme/v1/seo/indexnow/push', method: 'POST', data: { urls: list } })
    } catch (e) {
      // silent fail; backend may not be ready
      console.error(e)
    }
  }

  // Templates settings panel (MVP) - use function declaration to hoist
  function TemplatesPanel () {
    const initial = MyThemeSEO.settings?.templates || {}
    const [tpls, setTpls] = useState({
      post: { title_template: initial.post?.title_template || '{title} - {site}', description_template: initial.post?.description_template || '{title} | {site}' },
      page: { title_template: initial.page?.title_template || '{title} - {site}', description_template: initial.page?.description_template || '{title} | {site}' },
      product: { title_template: initial.product?.title_template || '{title} - {site}', description_template: initial.product?.description_template || '{title} | {site}' },
      category: { title_template: initial.category?.title_template || '{title} - {site}', description_template: initial.category?.description_template || '{title} | {site}' },
      product_cat: { title_template: initial.product_cat?.title_template || '{title} - {site}', description_template: initial.product_cat?.description_template || '{title} | {site}' },
      homepage: { title_template: initial.homepage?.title_template || '{site}', description_template: initial.homepage?.description_template || '{site}' }
    })
    const [saving, setSaving] = useState(false)
    const [note, setNote] = useState(null)

    const update = (key, field, value) => {
      setTpls((cur) => ({ ...cur, [key]: { ...(cur[key] || {}), [field]: value } }))
    }

    const save = async () => {
      setSaving(true)
      setNote(null)
      try {
        await apiFetch({ path: '/mytheme/v1/seo/settings', method: 'POST', data: { settings: { templates: tpls } } })
        setNote({ status: 'success', message: __('Templates saved.', 'mytheme-seo') })
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('Save failed.', 'mytheme-seo') })
      } finally {
        setSaving(false)
      }
    }

    const renderGroup = (key, label) => h(Card, { key, className: 'mytheme-seo-card', style: { marginBottom: '8px' } },
      h(CardHeader, null, label),
      h(CardBody, null,
        h(TextControl, { label: __('Title 模板', 'mytheme-seo'), value: tpls[key].title_template, onChange: (v) => update(key, 'title_template', v), help: __('变量：{title} {site} {locale}；分类可用 {term}；商品可用 {brand} {price}', 'mytheme-seo'), __nextHasNoMarginBottom: true }),
        h(TextareaControl, { label: __('Description 模板', 'mytheme-seo'), rows: 2, value: tpls[key].description_template, onChange: (v) => update(key, 'description_template', v), help: __('变量：{title} {site} {locale}；分类可用 {term}；商品可用 {brand} {price}', 'mytheme-seo'), __nextHasNoMarginBottom: true })
      )
    )

    return h(Card, { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('Templates（标题/描述模板）', 'mytheme-seo')),
      h(CardBody, null,
        renderGroup('post', 'Post'),
        renderGroup('page', 'Page'),
        renderGroup('product', 'Product'),
        renderGroup('category', 'Category'),
        renderGroup('product_cat', 'Product Category'),
        renderGroup('homepage', 'Homepage'),
        h(Button, { variant: 'primary', onClick: save, disabled: saving, style: { marginTop: '8px' } }, saving ? __('Saving…', 'mytheme-seo') : __('保存模板', 'mytheme-seo')),
        note && h(Notice, { status: note.status, onRemove: () => setNote(null), isDismissible: true }, note.message)
      )
    )
  }
  const emptyImageEntry = () => ({
    url: '',
    alt: '',
    title: '',
    caption: '',
    focus_keyword: '',
    license: '',
    creator: '',
    credit_url: '',
    active: true
  })

  const normalizeImageEntry = (entry) => {
    if (!entry || typeof entry !== 'object') {
      return emptyImageEntry()
    }
    const base = emptyImageEntry()
    return {
      ...base,
      ...entry,
      active: entry.active === undefined ? true : Boolean(entry.active)
    }
  }

  const IndexNowPanel = () => {
    const [enabled, setEnabled] = useState(Boolean(MyThemeSEO.settings?.indexnow?.enabled))
    const [key, setKey] = useState(String(MyThemeSEO.settings?.indexnow?.key || ''))
    const [note, setNote] = useState(null)
    const [testingUrl, setTestingUrl] = useState('')
    const [pushAllLocales, setPushAllLocales] = useState(Boolean(MyThemeSEO.settings?.indexnow?.pushAllLocales ?? true))
    const [defaultNoPrefix, setDefaultNoPrefix] = useState(Boolean(MyThemeSEO.settings?.indexnow?.defaultNoPrefix ?? true))
    const [tplPage, setTplPage] = useState(MyThemeSEO.settings?.indexnow?.tplPage || '/{locale}/{slug}')
    const [tplPost, setTplPost] = useState(MyThemeSEO.settings?.indexnow?.tplPost || '/{locale}/{slug}')
    const [tplProduct, setTplProduct] = useState(MyThemeSEO.settings?.indexnow?.tplProduct || '/{locale}/product/{slug}')
    const [tplCategory, setTplCategory] = useState(MyThemeSEO.settings?.indexnow?.tplCategory || '/{locale}/category/{slug}')
    const [tplProductCat, setTplProductCat] = useState(MyThemeSEO.settings?.indexnow?.tplProductCat || '/{locale}/product-category/{slug}')
    const site = (typeof window !== 'undefined' ? window.location.origin : '')
    const keyLocation = key ? `${site}/${key}.txt` : ''

    const saveSettings = async (next) => {
      try {
        await apiFetch({ path: '/mytheme/v1/seo/settings', method: 'POST', data: { settings: { indexnow: next } } })
        setNote({ status: 'success', message: __('IndexNow 设置已保存。', 'mytheme-seo') })
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('保存失败。', 'mytheme-seo') })
      }
    }

    const generateKey = () => {
      const rnd = Array.from(crypto.getRandomValues(new Uint8Array(24))).map((b) => b.toString(16).padStart(2, '0')).join('')
      setKey(rnd)
    }

    const verifyKeyFile = async () => {
      if (!key) { setNote({ status: 'error', message: __('请先生成或填写 Key。', 'mytheme-seo') }); return }
      try {
        const res = await fetch(`${site}/${key}.txt`, { method: 'GET', cache: 'no-cache' })
        if (res.ok) setNote({ status: 'success', message: __('Key 文件可访问。', 'mytheme-seo') })
        else setNote({ status: 'error', message: __('无法访问 Key 文件，请确认已部署到站点根目录。', 'mytheme-seo') })
      } catch (e) {
        setNote({ status: 'error', message: __('验证失败，请检查网络或跨域。', 'mytheme-seo') })
      }
    }

    const saveAll = () => saveSettings({ enabled, key, keyLocation: key ? `${site}/${key}.txt` : '', pushAllLocales, defaultNoPrefix, tplPage, tplPost, tplProduct, tplCategory, tplProductCat })

    // Preview-by-IDs state
    const [pvType, setPvType] = useState('post')
    const [pvId, setPvId] = useState('')
    const [pvTax, setPvTax] = useState('category')
    const [pvLocales, setPvLocales] = useState('')
    const [pvLoading, setPvLoading] = useState(false)
    const [pvResult, setPvResult] = useState([])

    const previewByIds = async () => {
      const idNum = parseInt(pvId, 10)
      if (!idNum && pvType !== 'homepage') { setNote({ status: 'error', message: __('请输入有效 ID（首页可不填）。', 'mytheme-seo') }); return }
      setPvLoading(true)
      setPvResult([])
      try {
        const locales = (pvLocales || '').split(/[,\s]+/).filter(Boolean)
        const data = { type: pvType, id: pvType === 'homepage' ? 0 : idNum, taxonomy: pvType === 'taxonomy' ? pvTax : undefined, locales }
        const res = await apiFetch({ path: '/mytheme/v1/seo/indexnow/preview-ids', method: 'POST', data })
        const urls = Array.isArray(res?.urls) ? res.urls : []
        setPvResult(urls)
        if (!urls.length) setNote({ status: 'warning', message: __('未生成任何 URL，请检查模板或对象/语言。', 'mytheme-seo') })
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('预览失败，请稍后再试。', 'mytheme-seo') })
      } finally {
        setPvLoading(false)
      }
    }

    return h(Card, { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('搜索引擎集成（IndexNow）', 'mytheme-seo')),
      h(CardBody, null,
        h(CheckboxControl, { label: __('启用 IndexNow', 'mytheme-seo'), checked: enabled, onChange: setEnabled }),
        h(NextText, { label: __('Key'), value: key, onChange: setKey, help: __('将同名文件 <key>.txt 放到站点根目录，文件内容为此 Key。', 'mytheme-seo') }),
        key && h('p', null, __('Key 文件路径：', 'mytheme-seo'), h('code', null, `${site}/${key}.txt`)),
        h('div', null,
          h(Button, { variant: 'secondary', onClick: generateKey, style: { marginRight: '8px' } }, __('生成 Key', 'mytheme-seo')),
          h(Button, { variant: 'secondary', onClick: verifyKeyFile, style: { marginRight: '8px' } }, __('验证 Key 文件', 'mytheme-seo')),
          h(Button, { variant: 'primary', onClick: saveAll }, __('保存设置', 'mytheme-seo'))
        ),
        h('hr'),
        h(CheckboxControl, { label: __('保存后推送全部语言', 'mytheme-seo'), checked: pushAllLocales, onChange: setPushAllLocales }),
        h(CheckboxControl, { label: __('默认语言不加前缀', 'mytheme-seo'), checked: defaultNoPrefix, onChange: setDefaultNoPrefix }),
        h(TextControl, { label: __('Page 模板'), value: tplPage, onChange: setTplPage, help: __('示例：/{locale}/{slug}', 'mytheme-seo') }),
        h(TextControl, { label: __('Post 模板'), value: tplPost, onChange: setTplPost, help: __('示例：/{locale}/{slug} 或 /{locale}/blog/{slug}', 'mytheme-seo') }),
        h(TextControl, { label: __('Product 模板'), value: tplProduct, onChange: setTplProduct, help: __('示例：/{locale}/product/{slug}', 'mytheme-seo') }),
        h(TextControl, { label: __('Category 模板'), value: tplCategory, onChange: setTplCategory, help: __('示例：/{locale}/category/{slug}', 'mytheme-seo') }),
        h(TextControl, { label: __('Product Category 模板'), value: tplProductCat, onChange: setTplProductCat, help: __('示例：/{locale}/product-category/{slug}', 'mytheme-seo') }),
        h('p', { className: 'mytheme-seo-tip' }, __('若某语言存在 canonical，将优先使用 canonical。', 'mytheme-seo')),
        h(Button, { variant: 'primary', onClick: saveAll, style: { marginBottom: '8px' } }, __('保存 URL 规则', 'mytheme-seo')),
        h('hr'),
        h(NextText, { label: __('测试推送 URL', 'mytheme-seo'), value: testingUrl, onChange: setTestingUrl }),
        h(Button, { variant: 'secondary', onClick: async () => { await pushIndexNow([testingUrl]) } }, __('立即测试推送', 'mytheme-seo')),
        h('hr'),
        h('h4', null, __('预览按对象推送将生成的 URL（后端解析）', 'mytheme-seo')),
        h(NextSelect, { label: __('对象类型', 'mytheme-seo'), value: pvType, onChange: setPvType, options: [
          { label: 'Post', value: 'post' },
          { label: 'Page', value: 'page' },
          { label: 'Product', value: 'product' },
          { label: 'Taxonomy', value: 'taxonomy' },
          { label: 'Homepage', value: 'homepage' }
        ]}),
        pvType !== 'homepage' && h(NextText, { label: __('对象 ID（Homepage 可留空或 0）', 'mytheme-seo'), value: pvId, onChange: setPvId }),
        pvType === 'taxonomy' && h(NextSelect, { label: __('Taxonomy'), value: pvTax, onChange: setPvTax, options: [
          { label: 'category', value: 'category' },
          { label: 'product_cat', value: 'product_cat' }
        ]}),
        h(NextText, { label: __('Locales（逗号或空格分隔；留空按设置）', 'mytheme-seo'), value: pvLocales, onChange: setPvLocales,
          help: __('示例：en zh fr；留空表示按“保存后推送全部语言”设置决定。', 'mytheme-seo') }),
        h(Button, { variant: 'secondary', onClick: previewByIds, disabled: pvLoading, style: { marginBottom: '8px' } }, pvLoading ? __('预览中…', 'mytheme-seo') : __('预览 URL', 'mytheme-seo')),
        pvResult.length > 0 && h('div', { style: { border: '1px solid #dcdcde', borderRadius: '6px', padding: '8px', marginTop: '6px' } },
          h('div', { style: { marginBottom: '8px', display: 'flex', gap: '8px' } },
            h(Button, { variant: 'secondary', onClick: () => copyToClipboard(pvResult.join('\n')) }, __('复制全部', 'mytheme-seo')),
            h(Button, { variant: 'primary', onClick: bulkPushPreview }, __('批量推送这些 URL', 'mytheme-seo'))
          ),
          pvResult.map((u, i) => h('div', { key: i, style: { wordBreak: 'break-all' } }, u))
        ),
        note && h(Notice, { status: note.status, onRemove: () => setNote(null), isDismissible: true }, note.message)
      )
    )
  }

  const IndexNowLogsPanel = () => {
    const [logFrom, setLogFrom] = useState('')
    const [logTo, setLogTo] = useState('')
    const [logType, setLogType] = useState('any')
    const [logStatus, setLogStatus] = useState('any')
    const [logLoading, setLogLoading] = useState(false)
    const [logItems, setLogItems] = useState([])
    const [note, setNote] = useState(null)

    const loadLogs = async () => {
      setLogLoading(true)
      setLogItems([])
      try {
        const qs = new URLSearchParams()
        if (logFrom) qs.set('from', logFrom)
        if (logTo) qs.set('to', logTo)
        if (logType !== 'any') qs.set('type', logType)
        if (logStatus !== 'any') qs.set('status', logStatus)
        const res = await apiFetch({ path: `/mytheme/v1/seo/indexnow/logs?${qs.toString()}` })
        setLogItems(Array.isArray(res?.items) ? res.items : [])
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('加载日志失败。', 'mytheme-seo') })
      } finally {
        setLogLoading(false)
      }
    }

    const retryLog = async (id) => {
      try {
        await apiFetch({ path: `/mytheme/v1/seo/indexnow/retry/${id}`, method: 'POST' })
        setNote({ status: 'success', message: __('已加入重试队列。', 'mytheme-seo') })
        await loadLogs()
      } catch (e) {
        console.error(e)
        setNote({ status: 'error', message: __('重试失败。', 'mytheme-seo') })
      }
    }

    const exportCSV = () => {
      const headers = ['id','type','status','locales','urls','object_id','taxonomy','created_at','attempts']
      const rows = (logItems || []).map((it) => [
        it.id,
        it.type || '',
        it.status || '',
        Array.isArray(it.locales) ? it.locales.join('|') : (it.locales || ''),
        Array.isArray(it.urls) ? it.urls.join('|') : (it.url || ''),
        it.object_id ?? '',
        it.taxonomy ?? '',
        it.created_at || '',
        it.attempts ?? ''
      ])
      const csv = [headers.join(','), ...rows.map((r) => r.map((c) => '"' + String(c).replace(/"/g,'""') + '"').join(','))].join('\n')
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `indexnow-logs-${Date.now()}.csv`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    }

    return h(Card, { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('IndexNow 日志', 'mytheme-seo')),
      h(CardBody, null,
        h('div', { style: { display: 'grid', gridTemplateColumns: '1fr 1fr 1fr 1fr auto auto', gap: '8px', alignItems: 'end', marginBottom: '8px' } },
          h(NextText, { label: __('开始日期 (YYYY-MM-DD)', 'mytheme-seo'), value: logFrom, onChange: setLogFrom }),
          h(NextText, { label: __('结束日期 (YYYY-MM-DD)', 'mytheme-seo'), value: logTo, onChange: setLogTo }),
          h(NextSelect, { label: __('对象类型', 'mytheme-seo'), value: logType, onChange: setLogType, options: [
            { label: __('全部', 'mytheme-seo'), value: 'any' },
            { label: 'Post', value: 'post' },
            { label: 'Page', value: 'page' },
            { label: 'Product', value: 'product' },
            { label: 'Taxonomy', value: 'taxonomy' },
            { label: 'Homepage', value: 'homepage' }
          ]}),
          h(NextSelect, { label: __('状态', 'mytheme-seo'), value: logStatus, onChange: setLogStatus, options: [
            { label: __('全部', 'mytheme-seo'), value: 'any' },
            { label: __('成功', 'mytheme-seo'), value: 'success' },
            { label: __('失败', 'mytheme-seo'), value: 'error' }
          ]}),
          h(Button, { variant: 'secondary', onClick: loadLogs, disabled: logLoading }, logLoading ? __('加载中…', 'mytheme-seo') : __('加载日志', 'mytheme-seo')),
          h(Button, { variant: 'secondary', onClick: exportCSV, disabled: !logItems.length }, __('导出 CSV', 'mytheme-seo'))
        ),
        logItems.length > 0 && h('div', { style: { border: '1px solid #dcdcde', borderRadius: '6px', padding: '8px' } },
          logItems.map((it) => h('div', { key: it.id, style: { display: 'flex', gap: '8px', alignItems: 'center', marginBottom: '6px' } },
            h('span', { style: { minWidth: '84px' } }, it.type || '—'),
            h('span', { style: { color: it.status === 'success' ? 'green' : (it.status === 'error' ? 'red' : 'inherit') } }, it.status || '—'),
            h('span', { style: { flex: 1, wordBreak: 'break-all' } }, Array.isArray(it.urls) ? it.urls.join(', ') : (it.url || '')),
            h('span', null, it.locales ? (Array.isArray(it.locales) ? it.locales.join(',') : it.locales) : ''),
            it.status === 'error' && h(Button, { variant: 'secondary', onClick: () => retryLog(it.id) }, __('重试', 'mytheme-seo'))
          ))
        ),
        note && h(Notice, { status: note.status, onRemove: () => setNote(null), isDismissible: true }, note.message)
      )
    )
  }

  const TaxonomySeoEditor = ({ taxonomy, languages, currentLocale }) => {
    const [termIdInput, setTermIdInput] = useState('')
    const [searchQuery, setSearchQuery] = useState('')
    const [searching, setSearching] = useState(false)
    const [results, setResults] = useState([])
    const [payload, setPayload] = useState({})
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [notice, setNotice] = useState(null)
    const [selectedTerm, setSelectedTerm] = useState(null)

    const normalizedPayload = useMemo(() => normalizePayload(payload, languages), [payload, languages])

    const taxLabel = taxonomy === 'product_cat' ? __('产品分类 SEO', 'mytheme-seo') : __('分类 SEO', 'mytheme-seo')
    const apiPathFor = (id) => `/mytheme/v1/seo/taxonomy/${taxonomy}/${id}`
    const wpRouteForTax = () => (taxonomy === 'product_cat' ? '/wp/v2/product_cat' : '/wp/v2/categories')

    const fetchSeo = async () => {
      const termId = parseInt(termIdInput, 10)
      if (!termId) {
        setNotice({ status: 'error', message: __('Enter a valid numeric term ID.', 'mytheme-seo') })
        return
      }
      setLoading(true)
      setNotice(null)
      try {
        const response = await apiFetch({ path: apiPathFor(termId) })
        setPayload(response.payload || {})
      } catch (error) {
        console.error(error)
        setNotice({ status: 'error', message: __('Failed to load taxonomy SEO payload.', 'mytheme-seo') })
      } finally {
        setLoading(false)
      }
    }

    const handleSave = async () => {
      const termId = parseInt(termIdInput, 10)
      if (!termId) {
        setNotice({ status: 'error', message: __('Enter a valid numeric term ID before saving.', 'mytheme-seo') })
        return
      }
      setSaving(true)
      setNotice(null)
      try {
        await apiFetch({ path: apiPathFor(termId), method: 'POST', data: { payload: normalizedPayload } })
        setNotice({ status: 'success', message: __('Taxonomy SEO saved.', 'mytheme-seo') })
        const settings = MyThemeSEO.settings?.indexnow || {}
        if (settings.enabled) {
          const pushAll = Boolean(settings.pushAllLocales ?? true)
          const locales = pushAll ? languages : [activeLocale]
          await pushIndexNowIds({ type: 'taxonomy', taxonomy, id: termId, locales })
        }
      } catch (error) {
        console.error(error)
        setNotice({ status: 'error', message: __('Failed to save taxonomy SEO.', 'mytheme-seo') })
      } finally {
        setSaving(false)
      }
    }

    const searchTerms = async () => {
      setSearching(true)
      setResults([])
      try {
        const route = wpRouteForTax()
        const qs = new URLSearchParams({ search: searchQuery || '', per_page: '20' })
        const items = await apiFetch({ path: `${route}?${qs.toString()}` })
        const mapped = (Array.isArray(items) ? items : []).map((it) => ({ id: it.id, title: it.name || `(ID ${it.id})` }))
        setResults(mapped)
      } catch (e) {
        console.error(e)
        setResults([])
      } finally {
        setSearching(false)
      }
    }

    const updateLocale = (locale, value) => {
      setPayload((current) => ({
        ...(current || {}),
        [locale]: value
      }))
    }

    const hasLanguages = languages.length > 0
    const activeLocale = hasLanguages ? (currentLocale && languages.includes(currentLocale) ? currentLocale : languages[0]) : ''
    const activeValue = activeLocale ? normalizedPayload[activeLocale] || emptyPayload() : emptyPayload()

    return h(
      Card,
      { className: 'mytheme-seo-card' },
      h(CardHeader, null, taxLabel),
      h(
        CardBody,
        null,
        h(TextControl, {
          label: __('Term ID', 'mytheme-seo'),
          value: termIdInput,
          onChange: setTermIdInput,
          help: __('Select a term via search or enter ID to fetch.', 'mytheme-seo')
        }),
        h(TextControl, {
          label: __('Search term', 'mytheme-seo'),
          value: searchQuery,
          onChange: setSearchQuery
        }),
        h(
          Button,
          { variant: 'secondary', onClick: searchTerms, disabled: searching },
          searching ? __('Searching…', 'mytheme-seo') : __('Search', 'mytheme-seo')
        ),
        results.length > 0 &&
          h(
            'div',
            { style: { marginTop: '12px', border: '1px solid #dcdcde', borderRadius: '6px', padding: '8px' } },
            results.map((r) =>
              h(
                Button,
                {
                  key: r.id,
                  variant: 'secondary',
                  onClick: async () => { setSelectedTerm(r); setTermIdInput(String(r.id)); await fetchSeo() },
                  style: { display: 'block', width: '100%', textAlign: 'left', marginBottom: '6px' }
                },
                `#${r.id} — ${r.title}`
              )
            )
          ),
        h(
          Button,
          { variant: 'secondary', onClick: fetchSeo, disabled: !languages.length || loading },
          loading ? __('Loading…', 'mytheme-seo') : __('Fetch SEO data', 'mytheme-seo')
        ),
        notice &&
          h(Notice, { status: notice.status, onRemove: () => setNotice(null), isDismissible: true }, notice.message),
        loading && h(Spinner, null),
        !loading && !hasLanguages &&
          h('p', { className: 'mytheme-seo-empty' }, __('Add locales before editing payload.', 'mytheme-seo')),
        !loading && hasLanguages &&
          h(Fragment, null,
            h(LocaleForm, { locale: activeLocale, value: activeValue, onChange: updateLocale }),
            h('div', { style: { marginTop: '8px' } },
              h(Button, { variant: 'secondary', onClick: () => {
                const site = (typeof window !== 'undefined' ? window.location.hostname : '')
                const key = taxonomy === 'product_cat' ? 'product_cat' : 'category'
                const tplCfg = (MyThemeSEO.settings?.templates || {})[key] || {}
                const ctx = { title: activeValue.title || '', site, locale: activeLocale, term: (selectedTerm && selectedTerm.title) || '' }
                const next = { ...activeValue }
                if (tplCfg.title_template) next.title = renderTemplate(tplCfg.title_template, ctx)
                if (tplCfg.description_template) next.description = renderTemplate(tplCfg.description_template, ctx)
                updateLocale(activeLocale, next)
              } }, __('按模板回填（当前语言）', 'mytheme-seo')),
              h(Button, { variant: 'secondary', style: { marginLeft: '8px' }, onClick: () => {
                const site = (typeof window !== 'undefined' ? window.location.hostname : '')
                const key = taxonomy === 'product_cat' ? 'product_cat' : 'category'
                const tplCfg = (MyThemeSEO.settings?.templates || {})[key] || {}
                setPayload((cur) => {
                  const base = normalizePayload(cur || {}, languages)
                  const out = { ...(cur || {}) }
                  languages.forEach((loc) => {
                    const ctx = { title: base[loc]?.title || '', site, locale: loc, term: (selectedTerm && selectedTerm.title) || '' }
                    const val = { ...(base[loc] || emptyPayload()) }
                    if (tplCfg.title_template) val.title = renderTemplate(tplCfg.title_template, ctx)
                    if (tplCfg.description_template) val.description = renderTemplate(tplCfg.description_template, ctx)
                    out[loc] = val
                  })
                  return out
                })
              } }, __('按模板回填（全部语言）', 'mytheme-seo'))
            )
          ),
        h(
          Button,
          { variant: 'primary', onClick: handleSave, disabled: saving || !hasLanguages },
          saving ? __('Saving…', 'mytheme-seo') : __('Save taxonomy SEO', 'mytheme-seo')
        )
      )
    )
  }

  const mythemeEmptyVideoEntry = () => ({
    title: '',
    description: '',
    focus_keyword: '',
    url: '',
    embed_url: '',
    content_url: '',
    thumbnail_url: '',
    upload_date: '',
    duration: '',
    type: '',
    active: true
  })

  const normalizeVideoEntry = (entry) => {
    if (!entry || typeof entry !== 'object') {
      return mythemeEmptyVideoEntry()
    }
    const base = mythemeEmptyVideoEntry()
    return {
      ...base,
      ...entry,
      active: entry.active === undefined ? true : Boolean(entry.active)
    }
  }

  const DEFAULT_LOCALE_ENTRIES = [
    { code: 'en', label: 'English' },
    { code: 'fr', label: 'Français' },
    { code: 'de', label: 'Deutsch' },
    { code: 'es', label: 'Español' },
    { code: 'ja', label: '日本語' },
    { code: 'ko', label: '한국어' },
    { code: 'it', label: 'Italiano' },
    { code: 'pt', label: 'Português' },
    { code: 'ru', label: 'Русский' },
    { code: 'ar', label: 'العربية' },
    { code: 'fi', label: 'Suomi' },
    { code: 'da', label: 'Dansk' },
    { code: 'th', label: 'ไทย' },
    { code: 'sv', label: 'Svenska' },
    { code: 'id', label: 'Bahasa Indonesia' },
    { code: 'ms', label: 'Bahasa Melayu' },
    { code: 'be', label: 'Беларуская' },
    { code: 'tr', label: 'Türkçe' },
    { code: 'bn', label: 'বাংলা' },
    { code: 'fa', label: 'فارسی' },
    { code: 'nl', label: 'Nederlands' },
    { code: 'hi', label: 'हिन्दी' },
    { code: 'ur', label: 'اردو' },
    { code: 'mr', label: 'मराठी' },
    { code: 'pcm', label: 'Nigerian Pidgin' },
    { code: 'fil', label: 'Filipino' },
    { code: 'te', label: 'తెలుగు' },
    { code: 'ha', label: 'Hausa' },
    { code: 'ps', label: 'پښتو' },
    { code: 'sw', label: 'Kiswahili' },
    { code: 'tl', label: 'Tagalog' },
    { code: 'ta', label: 'தமிழ்' },
    { code: 'jv', label: 'Basa Jawa' },
    { code: 'zh', label: '中文' }
  ]

  const defaultLanguageCodes = DEFAULT_LOCALE_ENTRIES.map((entry) => entry.code)

  const localeLabels = DEFAULT_LOCALE_ENTRIES.reduce((acc, entry) => {
    acc[entry.code.toLowerCase()] = entry.label
    return acc
  }, {})

  const getLocaleLabel = (code) => {
    const key = typeof code === 'string' ? code.toLowerCase() : ''
    return localeLabels[key] || (typeof code === 'string' ? code.toUpperCase() : '')
  }

  const ImageListEditor = ({ images, onChange }) => {
    const list = Array.isArray(images) ? images : []

    const update = (index, field, value) => {
      const next = list.map((item, idx) => (idx === index ? { ...item, [field]: value } : item))
      onChange(next)
    }

    const toggleActive = (index, checked) => {
      update(index, 'active', checked)
    }

    const remove = (index) => {
      const next = list.filter((_, idx) => idx !== index)
      onChange(next)
    }

    const addNew = () => {
      onChange([...list, emptyImageEntry()])
    }

    return h(
      Fragment,
      null,
      h(
        Button,
        { variant: 'secondary', onClick: addNew, style: { marginBottom: '1rem' } },
        __('新增图片条目', 'mytheme-seo')
      ),
      list.length === 0 &&
        h('p', { className: 'mytheme-seo-empty' }, __('暂无图片条目。', 'mytheme-seo')),
      list.map((entry, index) =>
        h(
          Card,
          { key: index, className: 'mytheme-seo-image-card' },
          h(CardHeader, null, `${__('图片', 'mytheme-seo')} #${index + 1}`),
          h(
            CardBody,
            null,
            h(TextControl, {
              label: __('图片 URL', 'mytheme-seo'),
              value: entry.url,
              onChange: (v) => update(index, 'url', v)
            }),
            h(TextControl, {
              label: __('Alt 文本', 'mytheme-seo'),
              value: entry.alt,
              onChange: (v) => update(index, 'alt', v)
            }),
            h(TextControl, {
              label: __('Title 文本', 'mytheme-seo'),
              value: entry.title,
              onChange: (v) => update(index, 'title', v)
            }),
            h(TextareaControl, {
              label: __('Caption / 描述', 'mytheme-seo'),
              rows: 3,
              value: entry.caption,
              onChange: (v) => update(index, 'caption', v)
            }),
            h(TextControl, {
              label: __('图片 Focus Keyword', 'mytheme-seo'),
              value: entry.focus_keyword,
              onChange: (v) => update(index, 'focus_keyword', v)
            }),
            h(TextControl, {
              label: __('版权/许可证', 'mytheme-seo'),
              value: entry.license,
              onChange: (v) => update(index, 'license', v)
            }),
            h(TextControl, {
              label: __('摄影师/创作者', 'mytheme-seo'),
              value: entry.creator,
              onChange: (v) => update(index, 'creator', v)
            }),
            h(TextControl, {
              label: __('版权链接', 'mytheme-seo'),
              value: entry.credit_url,
              onChange: (v) => update(index, 'credit_url', v)
            }),
            h(CheckboxControl, {
              label: __('启用此图片条目', 'mytheme-seo'),
              checked: entry.active,
              onChange: (v) => toggleActive(index, v)
            }),
            h(
              Button,
              {
                variant: 'secondary',
                onClick: () => remove(index),
                style: { marginTop: '0.5rem' }
              },
              __('删除此条目', 'mytheme-seo')
            )
          )
        )
      )
    )
  }

  const VideoListEditor = ({ videos, onChange }) => {
    const list = Array.isArray(videos) ? videos : []

    const update = (index, field, value) => {
      const next = list.map((item, idx) => (idx === index ? { ...item, [field]: value } : item))
      onChange(next)
    }

    const remove = (index) => {
      const next = list.filter((_, idx) => idx !== index)
      onChange(next)
    }

    const addNew = () => {
      onChange([...list, mythemeEmptyVideoEntry()])
    }

    return h(
      Fragment,
      null,
      h(
        Button,
        { variant: 'secondary', onClick: addNew, style: { marginBottom: '1rem' } },
        __('新增视频条目', 'mytheme-seo')
      ),
      list.length === 0 &&
        h('p', { className: 'mytheme-seo-empty' }, __('暂无视频条目。', 'mytheme-seo')),
      list.map((entry, index) =>
        h(
          Card,
          { key: index, className: 'mytheme-seo-video-card' },
          h(CardHeader, null, `${__('视频', 'mytheme-seo')} #${index + 1}`),
          h(
            CardBody,
            null,
            h(TextControl, {
              label: __('视频标题', 'mytheme-seo'),
              value: entry.title,
              onChange: (v) => update(index, 'title', v)
            }),
            h(TextareaControl, {
              label: __('视频描述', 'mytheme-seo'),
              rows: 3,
              value: entry.description,
              onChange: (v) => update(index, 'description', v)
            }),
            h(TextControl, {
              label: __('视频 Focus Keyword', 'mytheme-seo'),
              value: entry.focus_keyword,
              onChange: (v) => update(index, 'focus_keyword', v)
            }),
            h(TextControl, {
              label: __('外部链接 / 页面 URL', 'mytheme-seo'),
              value: entry.url,
              onChange: (v) => update(index, 'url', v),
              help: __('例如 YouTube 页面链接，可选', 'mytheme-seo')
            }),
            h(TextControl, {
              label: __('嵌入 URL (iframe)', 'mytheme-seo'),
              value: entry.embed_url,
              onChange: (v) => update(index, 'embed_url', v),
              help: __('例如 https://www.youtube.com/embed/...', 'mytheme-seo')
            }),
            h(TextControl, {
              label: __('内容 URL (MP4)', 'mytheme-seo'),
              value: entry.content_url,
              onChange: (v) => update(index, 'content_url', v),
              help: __('自托管视频文件 URL，可选', 'mytheme-seo')
            }),
            h(TextControl, {
              label: __('缩略图 URL', 'mytheme-seo'),
              value: entry.thumbnail_url,
              onChange: (v) => update(index, 'thumbnail_url', v)
            }),
            h(TextControl, {
              label: __('上传日期', 'mytheme-seo'),
              value: entry.upload_date,
              onChange: (v) => update(index, 'upload_date', v),
              help: __('格式示例：2024-05-01 或 ISO8601', 'mytheme-seo')
            }),
            h(TextControl, {
              label: __('时长', 'mytheme-seo'),
              value: entry.duration,
              onChange: (v) => update(index, 'duration', v),
              help: __('建议使用 ISO 8601, 例如 PT2M30S', 'mytheme-seo')
            }),
            h(TextControl, {
              label: __('类型 (可选)', 'mytheme-seo'),
              value: entry.type,
              onChange: (v) => update(index, 'type', v),
              help: __('例如 product, tutorial, testimonial', 'mytheme-seo')
            }),
            h(CheckboxControl, {
              label: __('启用此视频条目', 'mytheme-seo'),
              checked: entry.active,
              onChange: (v) => update(index, 'active', v)
            }),
            h(
              Button,
              {
                variant: 'secondary',
                onClick: () => remove(index),
                style: { marginTop: '0.5rem' }
              },
              __('删除此条目', 'mytheme-seo')
            )
          )
        )
      )
    )
  }

  const ensureLanguagesUnique = (list) => {
    if (!Array.isArray(list)) {
      return []
    }
    const seen = new Set()
    return list.filter((entry) => {
      const value = String(entry || '').trim()
      if (!value) return false
      const key = value.toLowerCase()
      if (seen.has(key)) return false
      seen.add(key)
      return true
    })
  }

  const splitListInput = (value) => {
    if (Array.isArray(value)) {
      return value
    }
    if (value === undefined || value === null) {
      return []
    }
    return String(value)
      .split(/\r?\n|,|;/)
      .map((entry) => entry.trim())
      .filter(Boolean)
  }

  const ensureUniqueExact = (list) => {
    if (!Array.isArray(list)) {
      return []
    }
    const seen = new Set()
    return list.filter((entry) => {
      const value = String(entry || '').trim()
      if (!value) return false
      if (seen.has(value)) return false
      seen.add(value)
      return true
    })
  }

  const normalizeRobotsList = (value) => ensureUniqueExact(splitListInput(value))

  const emptyPayload = () => ({
    title: '',
    description: '',
    focus_keyword: '',
    images: [],
    video: [],
    og: { title: '', description: '', image: '' },
    twitter: { card: 'summary_large_image', title: '', description: '', image: '' },
    jsonld: {}
  })

  const normalizePayload = (payload, languages) => {
    const result = {}
    const langs = ensureLanguagesUnique(languages)
    langs.forEach((locale) => {
      const source = payload?.[locale]
      const base = emptyPayload()
      if (source && typeof source === 'object') {
        result[locale] = {
          ...base,
          ...source,
          og: { ...base.og, ...(source.og || {}) },
          twitter: { ...base.twitter, ...(source.twitter || {}) },
          jsonld: typeof source.jsonld === 'object' && source.jsonld !== null ? source.jsonld : {},
          images: Array.isArray(source.images) ? source.images.map(normalizeImageEntry) : [],
          video: Array.isArray(source.video) ? source.video.map(normalizeVideoEntry) : []
        }
      } else {
        result[locale] = base
      }
    })
    return result
  }

  const LocaleForm = ({ locale, value, onChange }) => {
    const update = (path, next) => {
      onChange(locale, {
        ...value,
        [path]: next
      })
    }

    const updateNested = (key, field, next) => {
      onChange(locale, {
        ...value,
        [key]: {
          ...value[key],
          [field]: next
        }
      })
    }

    return h(
      Panel,
      { className: 'mytheme-seo-locale-panel' },
      h(
        PanelBody,
        { title: `${locale}`.toUpperCase(), initialOpen: true },
        h(TextControl, {
          label: __('Meta title', 'mytheme-seo'),
          value: value.title,
          onChange: (v) => update('title', v)
        }),
        h(TextareaControl, {
          label: __('Meta description', 'mytheme-seo'),
          value: value.description,
          onChange: (v) => update('description', v)
        }),
        h(TextControl, {
          label: __('Focus keyword', 'mytheme-seo'),
          value: value.focus_keyword,
          onChange: (v) => update('focus_keyword', v),
          help: __('单个核心关键词，用于分析或关键词提示。', 'mytheme-seo')
        }),
        h('h4', null, __('Open Graph', 'mytheme-seo')),
        h(TextControl, {
          label: __('OG Title', 'mytheme-seo'),
          value: value.og.title,
          onChange: (v) => updateNested('og', 'title', v)
        }),
        h(TextareaControl, {
          label: __('OG Description', 'mytheme-seo'),
          value: value.og.description,
          onChange: (v) => updateNested('og', 'description', v)
        }),
        h(TextControl, {
          label: __('OG Image URL', 'mytheme-seo'),
          value: value.og.image,
          onChange: (v) => updateNested('og', 'image', v)
        }),
        h('h4', null, __('Twitter', 'mytheme-seo')),
        h(TextControl, {
          label: __('Card type', 'mytheme-seo'),
          value: value.twitter.card,
          onChange: (v) => updateNested('twitter', 'card', v),
          help: __('Common values: summary, summary_large_image', 'mytheme-seo')
        }),
        h(TextControl, {
          label: __('Twitter Title', 'mytheme-seo'),
          value: value.twitter.title,
          onChange: (v) => updateNested('twitter', 'title', v)
        }),
        h(TextareaControl, {
          label: __('Twitter Description', 'mytheme-seo'),
          value: value.twitter.description,
          onChange: (v) => updateNested('twitter', 'description', v)
        }),
        h(TextControl, {
          label: __('Twitter Image URL', 'mytheme-seo'),
          value: value.twitter.image,
          onChange: (v) => updateNested('twitter', 'image', v)
        }),
        h('h4', null, __('Custom JSON-LD (optional)', 'mytheme-seo')),
        h(TextareaControl, {
          help: __('Paste JSON. Leave empty to let Nuxt fallback to autogenerated schema.', 'mytheme-seo'),
          rows: 6,
          value: JSON.stringify(value.jsonld || {}, null, 2),
          onChange: (v) => {
            try {
              const parsed = v ? JSON.parse(v) : {}
              update('jsonld', parsed)
            } catch (err) {
              // Keep raw string when JSON invalid to avoid losing data
              update('jsonld', v)
            }
          }
        }),
        h('h4', null, __('Image SEO', 'mytheme-seo')),
        h(
          'p',
          { className: 'mytheme-seo-tip' },
          __('为文章或页面关键图片设置 Alt/Title、版权信息与 Focus Keyword。', 'mytheme-seo')
        ),
        h(ImageListEditor, {
          images: value.images,
          onChange: (list) => update('images', list)
        }),
        h('h4', null, __('Video SEO', 'mytheme-seo')),
        h(
          'p',
          { className: 'mytheme-seo-tip' },
          __('可为嵌入或自托管视频设置结构化数据、缩略图与焦点关键词。', 'mytheme-seo')
        ),
        h(VideoListEditor, {
          videos: value.video,
          onChange: (list) => update('video', list)
        })
      )
    )
  }

  const RobotsSettings = ({ robots, onSave, saving, notice }) => {
    const [routesInput, setRoutesInput] = useState((robots.noindex_routes || []).join('\n'))
    const [componentsInput, setComponentsInput] = useState((robots.noindex_components || []).join('\n'))
    const [userAgentsInput, setUserAgentsInput] = useState((robots.blocked_user_agents || []).join('\n'))

    useEffect(() => {
      setRoutesInput((robots.noindex_routes || []).join('\n'))
      setComponentsInput((robots.noindex_components || []).join('\n'))
      setUserAgentsInput((robots.blocked_user_agents || []).join('\n'))
    }, [robots])

    const handleSave = () => {
      const next = {
        noindex_routes: normalizeRobotsList(routesInput),
        noindex_components: normalizeRobotsList(componentsInput),
        blocked_user_agents: normalizeRobotsList(userAgentsInput)
      }
      onSave(next)
    }

    const applyDefaults = () => {
      const defaults = [
        '/cart',
        '/checkout/*',
        '/my-account/*',
        '/order-received',
        '/thank-you'
      ]
      setRoutesInput(defaults.join('\n'))
    }

    return h(
      Card,
      { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('Robots 控制', 'mytheme-seo')),
      h(
        CardBody,
        null,
        h(
          'p',
          null,
          __('为指定路由或组件输出 noindex。每行一个条目，支持通配符（例如 /cart/*）。', 'mytheme-seo')
        ),
        h(TextareaControl, {
          label: __('Noindex 路由列表', 'mytheme-seo'),
          rows: 6,
          value: routesInput,
          onChange: setRoutesInput,
          help: __('示例：/cart 或 /checkout/*', 'mytheme-seo')
        }),
        h(TextareaControl, {
          label: __('Noindex 组件 key 列表', 'mytheme-seo'),
          rows: 6,
          value: componentsInput,
          onChange: setComponentsInput,
          help: __('示例：CartSummaryBar、FloatingPromoBanner', 'mytheme-seo')
        }),
        h(TextareaControl, {
          label: __('屏蔽的爬虫 User-Agent', 'mytheme-seo'),
          rows: 6,
          value: userAgentsInput,
          onChange: setUserAgentsInput,
          help: __('每行一个 UA 关键字，例如: AhrefsBot、MJ12bot、PetalBot。保存到 settings.robots.blocked_user_agents。', 'mytheme-seo')
        }),
        h(
          Button,
          { variant: 'secondary', onClick: applyDefaults, style: { marginRight: '8px' } },
          __('填入常见购物流程 noindex', 'mytheme-seo')
        ),
        h(
          Button,
          { variant: 'primary', onClick: handleSave, disabled: saving },
          saving ? __('Saving…', 'mytheme-seo') : __('保存 Robots 设置', 'mytheme-seo')
        ),
        notice &&
          h(Notice, { status: notice.status, onRemove: notice.onDismiss, isDismissible: true }, notice.message)
      )
    )
  }

  const SeoEditor = ({ languages, currentLocale }) => {
    const [objectType, setObjectType] = useState('post') // post | page | product
    const [postIdInput, setPostIdInput] = useState('')
    const [searchQuery, setSearchQuery] = useState('')
    const [searching, setSearching] = useState(false)
    const [results, setResults] = useState([])
    const [listPage, setListPage] = useState(1)
    const [listTotal, setListTotal] = useState(0)
    const [selectedItem, setSelectedItem] = useState(null)
    const [payload, setPayload] = useState({})
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [notice, setNotice] = useState(null)

    // bulk editor state
    const [bulkField, setBulkField] = useState('title') // title | description
    const [bulkOp, setBulkOp] = useState('append') // append | prepend | replace
    const [bulkText, setBulkText] = useState('')
    const [bulkScope, setBulkScope] = useState('all') // all | current

    const normalizedPayload = useMemo(() => normalizePayload(payload, languages), [payload, languages])

    const apiPathFor = (type, id) => {
      if (type === 'product') return `/mytheme/v1/seo/product/${id}`
      // post/page 目前共用同一端点，根据后端实现需要可拆分
      return `/mytheme/v1/seo/${id}`
    }

    const wpRouteForType = (type) => {
      if (type === 'page') return '/wp/v2/pages'
      if (type === 'product') return '/wp/v2/search' // generic search for product without Woo REST creds
      return '/wp/v2/posts'
    }

    const searchObjects = async (page = 1) => {
      setSearching(true)
      setResults([])
      try {
        const route = wpRouteForType(objectType)
        const qsInit = { search: searchQuery || '', per_page: '20', page: String(page) }
        if (objectType === 'product') {
          qsInit.subtype = 'product'
        }
        const qs = new URLSearchParams(qsInit)
        const response = await apiFetch({ path: `${route}?${qs.toString()}` })
        const items = Array.isArray(response) ? response : []
        const mapped = items.map((it) => ({ id: it.id, title: (it.title && it.title.rendered) || it.title || it.name || `(ID ${it.id})` }))
        setResults(mapped)
        // Try to read total from headers if available via apiFetch lastResponse. Fallback to length * page.
        try {
          const last = apiFetch.lastResponse
          const total = last && last.headers && last.headers.get ? parseInt(last.headers.get('X-WP-Total') || '0', 10) : 0
          setListTotal(Number.isFinite(total) ? total : mapped.length)
        } catch (_) {
          setListTotal(mapped.length)
        }
        setListPage(page)
      } catch (e) {
        console.error(e)
        setResults([])
      } finally {
        setSearching(false)
      }
    }

    const fetchSeo = async () => {
      const postId = parseInt(postIdInput, 10)
      if (!postId) {
        setNotice({ status: 'error', message: __('Enter a valid numeric post ID.', 'mytheme-seo') })
        return
      }
      setLoading(true)
      setNotice(null)
      try {
        const response = await apiFetch({ path: apiPathFor(objectType, postId) })
        setPayload(response.payload || {})
      } catch (error) {
        console.error(error)
        setNotice({ status: 'error', message: __('Failed to load SEO payload. Check console for details.', 'mytheme-seo') })
      } finally {
        setLoading(false)
      }
    }

    const handleSave = async () => {
      const postId = parseInt(postIdInput, 10)
      if (!postId) {
        setNotice({ status: 'error', message: __('Enter a valid numeric post ID before saving.', 'mytheme-seo') })
        return
      }
      setSaving(true)
      setNotice(null)
      try {
        await apiFetch({
          path: apiPathFor(objectType, postId),
          method: 'POST',
          data: { payload: normalizedPayload }
        })
        setNotice({ status: 'success', message: __('SEO payload saved.', 'mytheme-seo') })
      } catch (error) {
        console.error(error)
        setNotice({ status: 'error', message: __('Failed to save SEO payload.', 'mytheme-seo') })
      } finally {
        setSaving(false)
      }
    }

    const updateLocale = (locale, value) => {
      setPayload((current) => ({
        ...(current || {}),
        [locale]: value
      }))
    }

    const hasLanguages = languages.length > 0
    const activeLocale = hasLanguages
      ? (currentLocale && languages.includes(currentLocale) ? currentLocale : languages[0])
      : ''
    const activeValue = activeLocale ? normalizedPayload[activeLocale] || emptyPayload() : emptyPayload()

    return h(
      Card,
      { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('编辑页面/文章 SEO 负载', 'mytheme-seo')),
      h(
        CardBody,
        null,
        h('div', { style: { display: 'flex', gap: '12px' } },
          // Left subpane: object list
          h('div', { style: { width: '220px', flexShrink: 0, position: 'sticky', top: '72px', alignSelf: 'flex-start' } },
            h(NextSelect, {
              label: __('对象类型', 'mytheme-seo'),
              value: objectType,
              options: [
                { label: 'Post', value: 'post' },
                { label: 'Page', value: 'page' },
                { label: 'Product', value: 'product' }
              ],
              onChange: (v) => { setObjectType(v); setListPage(1); searchObjects(1) }
            }),
            h(NextText, {
              label: __('搜索标题', 'mytheme-seo'),
              value: searchQuery,
              onChange: setSearchQuery
            }),
            h(Button, { variant: 'secondary', onClick: () => searchObjects(1), disabled: searching, style: { marginBottom: '8px' } }, searching ? __('搜索中…', 'mytheme-seo') : __('搜索', 'mytheme-seo')),
            h('div', { style: { maxHeight: '380px', overflow: 'auto', border: '1px solid #dcdcde', borderRadius: '6px' } },
              results.map((r) => h(Button, {
                key: r.id,
                variant: selectedItem && selectedItem.id === r.id ? 'primary' : 'secondary',
                onClick: async () => { setSelectedItem(r); setPostIdInput(String(r.id)); await fetchSeo() },
                style: { display: 'block', width: '100%', textAlign: 'left', borderRadius: 0 }
              }, `#${r.id} — ${r.title}`))
            ),
            h('div', { style: { display: 'flex', justifyContent: 'space-between', marginTop: '6px' } },
              h(Button, { variant: 'secondary', disabled: searching || listPage <= 1, onClick: () => searchObjects(listPage - 1) }, __('上一页', 'mytheme-seo')),
              h('span', null, `${listPage}`),
              h(Button, { variant: 'secondary', disabled: searching || results.length < 20, onClick: () => searchObjects(listPage + 1) }, __('下一页', 'mytheme-seo'))
            )
          ),
          // Right main editor
          h('div', { style: { flex: 1 } },
            h(NextSelect, {
              label: __('Object type', 'mytheme-seo'),
              value: objectType,
              options: [
                { label: 'Post', value: 'post' },
                { label: 'Page', value: 'page' },
                { label: 'Product', value: 'product' }
              ],
              onChange: setObjectType
            }),
            h(NextText, {
              label: __('WordPress 文章/页面 ID', 'mytheme-seo'),
              value: postIdInput,
              onChange: setPostIdInput,
              help: __('可在“文章/页面”列表悬停链接或通过 REST API 获取 ID。', 'mytheme-seo')
            }),
            h(
              Button,
              { variant: 'secondary', onClick: fetchSeo, disabled: !languages.length || loading },
              loading ? __('加载中…', 'mytheme-seo') : __('获取 SEO 数据', 'mytheme-seo')
            ),
            h('hr'),
            hasLanguages && h('h4', null, __('批量编辑器', 'mytheme-seo')),
            hasLanguages && h(NextSelect, { label: __('字段', 'mytheme-seo'), value: bulkField, onChange: setBulkField, options: [
              { label: 'Title', value: 'title' },
              { label: 'Description', value: 'description' }
            ]}),
            hasLanguages && h(NextSelect, { label: __('操作', 'mytheme-seo'), value: bulkOp, onChange: setBulkOp, options: [
              { label: __('追加', 'mytheme-seo'), value: 'append' },
              { label: __('前置', 'mytheme-seo'), value: 'prepend' },
              { label: __('替换', 'mytheme-seo'), value: 'replace' }
            ]}),
            hasLanguages && h(NextSelect, { label: __('作用范围', 'mytheme-seo'), value: bulkScope, onChange: setBulkScope, options: [
              { label: __('全部语言', 'mytheme-seo'), value: 'all' },
              { label: __('仅当前语言', 'mytheme-seo'), value: 'current' }
            ]}),
            hasLanguages && h(NextText, { label: __('文本', 'mytheme-seo'), value: bulkText, onChange: setBulkText }),
            hasLanguages && h(Button, { variant: 'secondary', onClick: () => {
              if (!bulkText && bulkOp !== 'replace') return
              setPayload((current) => {
                const next = { ...(current || {}) }
                const applyTo = bulkScope === 'all' ? languages : [activeLocale]
                applyTo.forEach((loc) => {
                  const base = normalizePayload(current || {}, languages)[loc] || emptyPayload()
                  const currentVal = String(base[bulkField] || '')
                  let newVal = currentVal
                  if (bulkOp === 'append') newVal = currentVal + bulkText
                  else if (bulkOp === 'prepend') newVal = bulkText + currentVal
                  else if (bulkOp === 'replace') newVal = bulkText
                  next[loc] = { ...base, [bulkField]: newVal }
                })
                return next
              })
            }, style: { marginBottom: '12px' } }, __('应用', 'mytheme-seo')),
            
            notice &&
              h(Notice, { status: notice.status, onRemove: () => setNotice(null), isDismissible: true }, notice.message),
            loading && h(Spinner, null),
            !loading && !hasLanguages &&
              h('p', { className: 'mytheme-seo-empty' }, __('Add locales before editing payload.', 'mytheme-seo')),
            !loading && hasLanguages &&
              h(Fragment, null,
                h(LocaleForm, { locale: activeLocale, value: activeValue, onChange: updateLocale }),
                h('div', { style: { marginTop: '8px' } },
                  h(Button, { variant: 'secondary', onClick: () => {
                    const site = (typeof window !== 'undefined' ? window.location.hostname : '')
                    const tplCfg = (MyThemeSEO.settings?.templates || {})[objectType] || {}
                    const { brand, price } = extractBrandPrice(activeValue.jsonld)
                    const ctx = { title: activeValue.title || '', site, locale: activeLocale, brand, price }
                    const next = { ...activeValue }
                    if (tplCfg.title_template) next.title = renderTemplate(tplCfg.title_template, ctx)
                    if (tplCfg.description_template) next.description = renderTemplate(tplCfg.description_template, ctx)
                    updateLocale(activeLocale, next)
                  }, style: { marginRight: '8px' } }, __('按模板回填（当前语言）', 'mytheme-seo')),
                  h(Button, { variant: 'secondary', onClick: () => {
                    const site = (typeof window !== 'undefined' ? window.location.hostname : '')
                    const tplCfg = (MyThemeSEO.settings?.templates || {})[objectType] || {}
                    setPayload((cur) => {
                      const base = normalizePayload(cur || {}, languages)
                      const out = { ...(cur || {}) }
                      languages.forEach((loc) => {
                        const { brand, price } = extractBrandPrice(base[loc]?.jsonld)
                        const ctx = { title: base[loc]?.title || '', site, locale: loc, brand, price }
                        const val = { ...(base[loc] || emptyPayload()) }
                        if (tplCfg.title_template) val.title = renderTemplate(tplCfg.title_template, ctx)
                        if (tplCfg.description_template) val.description = renderTemplate(tplCfg.description_template, ctx)
                        out[loc] = val
                      })
                      return out
                    })
                  } }, __('按模板回填（全部语言）', 'mytheme-seo'))
                )
              ),
            h(
              Button,
              { variant: 'primary', onClick: handleSave, disabled: saving || !hasLanguages },
              saving ? __('保存中…', 'mytheme-seo') : __('保存 SEO 负载', 'mytheme-seo')
            ),
            h(Button, { variant: 'secondary', style: { marginLeft: '8px' }, onClick: async () => {
              const postId = parseInt(postIdInput, 10)
              if (!postId) return
              await pushIndexNowIds({ type: objectType, id: postId, locales: [activeLocale] })
              setNotice({ status: 'success', message: __('已推送当前语言到 IndexNow。', 'mytheme-seo') })
            } }, __('推送当前语言', 'mytheme-seo')),
            h(Button, { variant: 'secondary', style: { marginLeft: '8px' }, onClick: async () => {
              const postId = parseInt(postIdInput, 10)
              if (!postId) return
              try {
                const settings = MyThemeSEO.settings?.indexnow || {}
                const pushAll = Boolean(settings.pushAllLocales ?? true)
                const locales = pushAll ? languages : [activeLocale]
                const res = await apiFetch({ path: '/mytheme/v1/seo/indexnow/preview-ids', method: 'POST', data: { type: objectType, id: postId, locales } })
                const urls = Array.isArray(res?.urls) ? res.urls : []
                alert(urls.length ? urls.join('\n') : __('未生成 URL，请检查模板与对象/语言设置。', 'mytheme-seo'))
              } catch (e) {
                console.error(e)
                setNotice({ status: 'error', message: __('预览失败，请稍后再试。', 'mytheme-seo') })
              }
            } }, __('预览 URL 列表', 'mytheme-seo'))
          )
        )
      )
    )
  }

  const HomepageSeoEditor = ({ languages, currentLocale }) => {
    const [payload, setPayload] = useState({})
    const [loading, setLoading] = useState(false)
    const [saving, setSaving] = useState(false)
    const [notice, setNotice] = useState(null)

    const normalizedPayload = useMemo(() => normalizePayload(payload, languages), [payload, languages])

    const fetchSeo = async () => {
      setLoading(true)
      setNotice(null)
      try {
        const res = await apiFetch({ path: '/mytheme/v1/seo/homepage' })
        setPayload(res.payload || {})
      } catch (e) {
        console.error(e)
        setNotice({ status: 'error', message: __('Failed to load homepage SEO.', 'mytheme-seo') })
      } finally {
        setLoading(false)
      }
    }

    const handleSave = async () => {
      setSaving(true)
      setNotice(null)
      try {
        await apiFetch({ path: '/mytheme/v1/seo/homepage', method: 'POST', data: { payload: normalizedPayload } })
        setNotice({ status: 'success', message: __('Homepage SEO saved.', 'mytheme-seo') })
        const settings = MyThemeSEO.settings?.indexnow || {}
        if (settings.enabled) {
          const pushAll = Boolean(settings.pushAllLocales ?? true)
          const locales = pushAll ? languages : [activeLocale]
          await pushIndexNowIds({ type: 'homepage', id: 0, locales })
        }
      } catch (e) {
        console.error(e)
        setNotice({ status: 'error', message: __('Failed to save homepage SEO.', 'mytheme-seo') })
      } finally {
        setSaving(false)
      }
    }

    const hasLanguages = languages.length > 0
    const activeLocale = hasLanguages ? (currentLocale && languages.includes(currentLocale) ? currentLocale : languages[0]) : ''
    const activeValue = activeLocale ? normalizedPayload[activeLocale] || emptyPayload() : emptyPayload()

    const updateLocale = (locale, value) => {
      setPayload((current) => ({ ...(current || {}), [locale]: value }))
    }

    useEffect(() => { fetchSeo() }, [])

    return h(
      Card,
      { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('首页 SEO', 'mytheme-seo')),
      h(
        CardBody,
        null,
        h(
          Button,
          { variant: 'secondary', onClick: fetchSeo, disabled: loading, style: { marginBottom: '8px' } },
          loading ? __('Loading…', 'mytheme-seo') : __('Refresh', 'mytheme-seo')
        ),
        notice && h(Notice, { status: notice.status, onRemove: () => setNotice(null), isDismissible: true }, notice.message),
        loading && h(Spinner, null),
        !loading && !hasLanguages && h('p', { className: 'mytheme-seo-empty' }, __('Add locales before editing payload.', 'mytheme-seo')),
        !loading && hasLanguages && h(Fragment, null,
          h(LocaleForm, { locale: activeLocale, value: activeValue, onChange: updateLocale }),
          h('div', { style: { marginTop: '8px' } },
            h(Button, { variant: 'secondary', onClick: () => {
              const site = (typeof window !== 'undefined' ? window.location.hostname : '')
              const tplCfg = (MyThemeSEO.settings?.templates || {}).homepage || {}
              const ctx = { title: activeValue.title || '', site, locale: activeLocale }
              const next = { ...activeValue }
              if (tplCfg.title_template) next.title = renderTemplate(tplCfg.title_template, ctx)
              if (tplCfg.description_template) next.description = renderTemplate(tplCfg.description_template, ctx)
              updateLocale(activeLocale, next)
            } }, __('按模板回填（当前语言）', 'mytheme-seo'))
          )
        ),
        h(
          Button,
          { variant: 'primary', onClick: handleSave, disabled: saving || !hasLanguages },
          saving ? __('Saving…', 'mytheme-seo') : __('Save SEO payload', 'mytheme-seo')
        ),
        h(Button, { variant: 'secondary', style: { marginLeft: '8px' }, onClick: async () => {
          try {
            const settings = MyThemeSEO.settings?.indexnow || {}
            const pushAll = Boolean(settings.pushAllLocales ?? true)
            const locales = pushAll ? languages : [activeLocale]
            const res = await apiFetch({ path: '/mytheme/v1/seo/indexnow/preview-ids', method: 'POST', data: { type: 'homepage', id: 0, locales } })
            const urls = Array.isArray(res?.urls) ? res.urls : []
            alert(urls.length ? urls.join('\n') : __('未生成 URL，请检查模板与对象/语言设置。', 'mytheme-seo'))
          } catch (e) {
            console.error(e)
            setNotice({ status: 'error', message: __('预览失败，请稍后再试。', 'mytheme-seo') })
          }
        } }, __('Preview URLs', 'mytheme-seo'))
      )
    )
  }

  const LogsPanel = () => {
    const [logs, setLogs] = useState([])
    const [loading, setLoading] = useState(false)
    const [notice, setNotice] = useState(null)

    const loadLogs = async () => {
      setLoading(true)
      setNotice(null)
      try {
        const res = await apiFetch({ path: '/mytheme/v1/seo/404-logs' })
        setLogs(Array.isArray(res?.logs) ? res.logs : (res || []))
      } catch (e) {
        console.error(e)
        setNotice({ status: 'error', message: __('Failed to load 404 logs.', 'mytheme-seo') })
      } finally {
        setLoading(false)
      }
    }

    const mutate = async (action, path) => {
      try {
        await apiFetch({ path: '/mytheme/v1/seo/404-logs', method: 'POST', data: { action, path } })
        if (action === 'clear_all') setLogs([])
        else setLogs((prev) => prev.filter((it) => it.path !== path))
      } catch (e) {
        console.error(e)
        setNotice({ status: 'error', message: __('Operation failed.', 'mytheme-seo') })
      }
    }

    useEffect(() => { loadLogs() }, [])

    return h(
      Card,
      { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('404 监测', 'mytheme-seo')),
      h(
        CardBody,
        null,
        h(
          Button,
          { variant: 'secondary', onClick: loadLogs, disabled: loading, style: { marginRight: '8px' } },
          loading ? __('Loading…', 'mytheme-seo') : __('Refresh', 'mytheme-seo')
        ),
        h(
          Button,
          { variant: 'secondary', onClick: () => mutate('clear_all') },
          __('Clear all', 'mytheme-seo')
        ),
        notice && h(Notice, { status: notice.status, onRemove: () => setNotice(null), isDismissible: true }, notice.message),
        loading && h(Spinner, null),
        !loading && logs && logs.length > 0 &&
          h(
            'table',
            { className: 'widefat striped', style: { marginTop: '12px' } },
            h('thead', null,
              h('tr', null,
                h('th', null, __('Path', 'mytheme-seo')),
                h('th', null, __('Count', 'mytheme-seo')),
                h('th', null, __('First Seen', 'mytheme-seo')),
                h('th', null, __('Last Seen', 'mytheme-seo')),
                h('th', null, __('Referrer', 'mytheme-seo')),
                h('th', null, __('Resolved', 'mytheme-seo')),
                h('th', null, __('Actions', 'mytheme-seo'))
              )
            ),
            h('tbody', null,
              logs.map((entry) =>
                h('tr', { key: entry.path },
                  h('td', null, entry.path),
                  h('td', null, entry.count || 0),
                  h('td', null, entry.first_seen ? new Date(entry.first_seen * 1000).toLocaleString() : ''),
                  h('td', null, entry.last_seen ? new Date(entry.last_seen * 1000).toLocaleString() : ''),
                  h('td', null, entry.last_referrer ? h('a', { href: entry.last_referrer, target: '_blank', rel: 'noopener noreferrer' }, __('查看', 'mytheme-seo')) : __('无', 'mytheme-seo')),
                  h('td', null, entry.resolved ? __('已解决', 'mytheme-seo') : __('未解决', 'mytheme-seo')),
                  h('td', null,
                    h(Button, { size: 'small', variant: 'secondary', onClick: () => mutate(entry.resolved ? 'mark_unresolved' : 'mark_resolved', entry.path), style: { marginRight: '6px' } }, entry.resolved ? __('标记未解决', 'mytheme-seo') : __('标记已解决', 'mytheme-seo')),
                    h(Button, { size: 'small', variant: 'secondary', onClick: () => mutate('delete', entry.path), style: { marginRight: '6px' } }, __('删除', 'mytheme-seo')),
                    h(Button, { size: 'small', variant: 'secondary', onClick: () => window.open(entry.path, '_blank') }, __('打开', 'mytheme-seo'))
                  )
                )
              )
            )
          ),
        !loading && (!logs || logs.length === 0) && h('p', { className: 'mytheme-seo-empty' }, __('No 404 logs yet.', 'mytheme-seo'))
      )
    )
  }

  const AuditList = ({ title, data, emptyMessage }) => {
    const count = data?.count || 0
    const items = Array.isArray(data?.items) ? data.items : []

    return h(Card, { className: 'mytheme-seo-card' },
      h(CardHeader, null, `${title}（${count}）`),
      h(CardBody, null,
        items.length === 0
          ? h('p', { style: { color: '#6b7280', margin: 0 } }, emptyMessage)
          : h('ul', {
              style: {
                listStyle: 'none',
                padding: 0,
                margin: 0,
                display: 'grid',
                gap: '8px'
              }
            },
            items.map((item) => h('li', {
              key: item.id,
              style: {
                border: '1px solid #e5e7eb',
                borderRadius: '6px',
                padding: '8px 10px',
                display: 'grid',
                gap: '4px'
              }
            },
            h('div', {
              style: {
                fontWeight: 600,
                color: '#111827'
              }
            }, item.title || __('(未命名)', 'mytheme-seo')),
            h('div', { style: { fontSize: '12px', color: '#6b7280' } }, `${__('类型', 'mytheme-seo')}: ${item.type || '-'}`),
            h('div', {
              style: {
                display: 'flex',
                gap: '8px',
                flexWrap: 'wrap'
              }
            },
            item.edit_url && h(Button, {
              variant: 'secondary',
              href: item.edit_url,
              target: '_blank'
            }, __('后台编辑', 'mytheme-seo')),
            item.view_url && h(Button, {
              variant: 'secondary',
              href: item.view_url,
              target: '_blank'
            }, __('前台预览', 'mytheme-seo'))
            )
            ))
          )
      )
    )
  }

  const SeoAuditPanel = () => {
    const [summary, setSummary] = useState({ products: null, content: null })
    const [missing, setMissing] = useState(null)
    const [loading, setLoading] = useState(true)
    const [testing, setTesting] = useState(false)
    const [error, setError] = useState(null)

    const loadSummary = async () => {
      setLoading(true)
      setError(null)
      try {
        const res = await apiFetch({ path: '/mytheme/v1/seo/audit' })
        setSummary({
          products: res?.products || { count: 0, items: [] },
          content: res?.content || { count: 0, items: [] }
        })
      } catch (e) {
        console.error(e)
        setError(__('加载 SEO 状态失败，请稍后再试。', 'mytheme-seo'))
      } finally {
        setLoading(false)
      }
    }

    useEffect(() => {
      loadSummary()
    }, [])

    const testMissing = async () => {
      setTesting(true)
      setError(null)
      try {
        const res = await apiFetch({ path: '/mytheme/v1/seo/audit?scope=missing' })
        setMissing(res?.missing || { count: 0, items: [] })
      } catch (e) {
        console.error(e)
        setError(__('测试未配置 SEO 的内容失败。', 'mytheme-seo'))
      } finally {
        setTesting(false)
      }
    }

    return h('div', {
      style: {
        marginTop: '16px',
        display: 'grid',
        gap: '16px'
      }
    },
    h(Card, { className: 'mytheme-seo-card' },
      h(CardHeader, null, __('SEO 状态总览', 'mytheme-seo')),
      h(CardBody, null,
        h('div', {
          style: {
            display: 'flex',
            gap: '8px',
            flexWrap: 'wrap'
          }
        },
        h(Button, {
          variant: 'secondary',
          onClick: loadSummary,
          disabled: loading
        }, loading ? __('加载中…', 'mytheme-seo') : __('刷新概览', 'mytheme-seo')),
        h(Button, {
          variant: 'primary',
          onClick: testMissing,
          disabled: testing
        }, testing ? __('测试中…', 'mytheme-seo') : __('测试未配置 SEO 的页面', 'mytheme-seo'))
        ),
        error && h(Notice, {
          status: 'error',
          onRemove: () => setError(null),
          isDismissible: true,
          style: { marginTop: '12px' }
        }, error),
        loading && h('div', { style: { marginTop: '12px' } }, h(Spinner, null)),
        !loading && h('div', {
          style: {
            marginTop: '16px',
            display: 'grid',
            gap: '16px'
          }
        },
        h(AuditList, {
          title: __('已配置 SEO 的商品', 'mytheme-seo'),
          data: summary.products,
          emptyMessage: __('暂无已配置 SEO 的商品。', 'mytheme-seo')
        }),
        h(AuditList, {
          title: __('已配置 SEO 的页面/文章', 'mytheme-seo'),
          data: summary.content,
          emptyMessage: __('暂无已配置 SEO 的页面或文章。', 'mytheme-seo')
        })
        ),
        missing && h(AuditList, {
          title: __('未配置 SEO 的内容', 'mytheme-seo'),
          data: missing,
          emptyMessage: __('很好！所有内容均已配置 SEO。', 'mytheme-seo')
        })
      )
    )
    )
  }

  const App = () => {
    const storedLanguages = ensureLanguagesUnique(MyThemeSEO.languages || [])
    const fallbackLanguages = storedLanguages.length ? storedLanguages : defaultLanguageCodes
    const [languages] = useState(fallbackLanguages)
    const [currentLocale, setCurrentLocale] = useState(() => (fallbackLanguages[0] || ''))

    const defaultRobots = MyThemeSEO.settings?.robots || { noindex_routes: [], noindex_components: [] }
    const [robots, setRobots] = useState({
      noindex_routes: normalizeRobotsList(defaultRobots.noindex_routes),
      noindex_components: normalizeRobotsList(defaultRobots.noindex_components)
    })
    const [savingRobots, setSavingRobots] = useState(false)
    const [robotsNotice, setRobotsNotice] = useState(null)

    const FEATURES = [
      { key: 'content', label: __('内容 SEO', 'mytheme-seo') },
      { key: 'category', label: __('分类 SEO', 'mytheme-seo') },
      { key: 'product_cat', label: __('产品分类 SEO', 'mytheme-seo') },
      { key: 'homepage', label: __('首页 SEO', 'mytheme-seo') },
      { key: 'templates', label: __('Templates', 'mytheme-seo') },
      { key: 'product_schema', label: __('Product Schema', 'mytheme-seo') },
      { key: 'product_schema_mapping', label: __('字段映射', 'mytheme-seo') },
      { key: 'sitemaps', label: __('Sitemaps', 'mytheme-seo') },
      { key: 'robots', label: __('Robots 控制', 'mytheme-seo') },
      { key: 'logs', label: __('404 监测', 'mytheme-seo') },
      { key: 'indexnow', label: __('搜索引擎集成', 'mytheme-seo') },
      { key: 'indexnow_logs', label: __('IndexNow 日志', 'mytheme-seo') }
    ]
    const [feature, setFeature] = useState('content')

    const handleSaveRobots = async (next) => {
      setSavingRobots(true)
      setRobotsNotice(null)
      try {
        await apiFetch({
          path: '/mytheme/v1/seo/settings',
          method: 'POST',
          data: { settings: { robots: next } }
        })
        setRobots(next)
        const tip = __('已保存 Robots 设置。请同步更新 robots.txt 或服务器 (Nginx/Apache) 的 UA 阻断配置，详见 tanzanite.md 的 robots.txt 章节。', 'mytheme-seo')
        setRobotsNotice({ status: 'success', message: tip, onDismiss: () => setRobotsNotice(null) })
      } catch (error) {
        console.error(error)
        setRobotsNotice({ status: 'error', message: __('保存失败，请稍后重试。', 'mytheme-seo'), onDismiss: () => setRobotsNotice(null) })
      } finally {
        setSavingRobots(false)
      }
    }

    const renderRightPane = () => {
      if (feature === 'robots') {
        return h(RobotsSettings, { robots, onSave: handleSaveRobots, saving: savingRobots, notice: robotsNotice })
      }
      if (feature === 'content') {
        return h(SeoEditor, { languages, currentLocale })
      }
      if (feature === 'category') {
        return h(TaxonomySeoEditor, { taxonomy: 'category', languages, currentLocale })
      }
      if (feature === 'product_cat') {
        return h(TaxonomySeoEditor, { taxonomy: 'product_cat', languages, currentLocale })
      }
      if (feature === 'homepage') {
        return h(HomepageSeoEditor, { languages, currentLocale })
      }
      if (feature === 'templates') {
        return h(TemplatesPanel)
      }
      if (feature === 'product_schema') {
        return h(ProductSchemaPanel)
      }
      if (feature === 'product_schema_mapping') {
        return h(ProductSchemaMappingPanel)
      }
      if (feature === 'sitemaps') {
        return h(SitemapsPanel)
      }
      if (feature === 'logs') {
        return h(LogsPanel)
      }
      if (feature === 'indexnow') {
        return h(IndexNowPanel)
      }
      if (feature === 'indexnow_logs') {
        return h(IndexNowLogsPanel)
      }
      return null
    }

    return h(
      'div',
      { className: 'mytheme-seo-locale-layout' },
      // Left: Languages
      h(
        'div',
        { className: 'mytheme-seo-locale-column' },
        h(Card, null,
          h(CardHeader, null, __('选择语言', 'mytheme-seo')),
          h(CardBody, null,
            languages.map((code) =>
              h(
                'button',
                {
                  key: code,
                  className: 'mytheme-seo-locale-item' + (code === currentLocale ? ' is-active' : ''),
                  onClick: () => setCurrentLocale(code)
                },
                h('span', { className: 'code' }, String(code || '').toUpperCase()),
                h('span', { className: 'label' }, getLocaleLabel(code))
              )
            )
          )
        )
      ),
      // Middle: Feature picker
      h(
        'div',
        { className: 'mytheme-seo-feature-column', style: { width: '200px', flexShrink: 0, alignSelf: 'flex-start', position: 'sticky', top: '0px' } },
        h(Card, null,
          h(CardHeader, null, __('功能', 'mytheme-seo')),
          h(CardBody, { className: 'mytheme-seo-feature-list' },
            FEATURES.map((f) =>
              h(
                Button,
                {
                  key: f.key,
                  variant: f.key === feature ? 'primary' : 'secondary',
                  className: 'mytheme-seo-feature-item' + (f.key === feature ? ' is-active' : ''),
                  onClick: () => setFeature(f.key),
                  style: { marginBottom: '8px' }
                },
                f.label
              )
            )
          )
        )
      ),
      // Right: Editor
      h(
        'div',
        { className: 'mytheme-seo-locale-content' },
        renderRightPane()
      )
    )
  }

if (document.getElementById('mytheme-seo-admin-app')) {
  wp.element.render(h(App), document.getElementById('mytheme-seo-admin-app'))
}

const auditRoot = document.getElementById('mytheme-seo-audit-root')
if (auditRoot) {
  wp.element.render(h(SeoAuditPanel), auditRoot)
}

})();
