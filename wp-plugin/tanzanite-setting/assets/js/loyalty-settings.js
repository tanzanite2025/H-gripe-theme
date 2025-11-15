(function () {
  console.log('[Tanzanite Loyalty] Script loaded, checking dependencies...');
  console.log('[Tanzanite Loyalty] wp:', typeof window.wp);
  console.log('[Tanzanite Loyalty] wp.element:', typeof (window.wp && window.wp.element));
  console.log('[Tanzanite Loyalty] wp.i18n:', typeof (window.wp && window.wp.i18n));
  
  if (!window.wp || !wp.element || !wp.i18n) {
    console.error('[Tanzanite Loyalty] Missing required dependencies!');
    return
  }
  
  console.log('[Tanzanite Loyalty] All dependencies loaded successfully');
  const { createElement: h, Fragment, useState, useEffect, useMemo, useCallback } = wp.element
  const { __ } = wp.i18n
  const components = wp.components || {}

  // Fallback primitives to avoid undefined React component types
  const Fallback = {
    TextControl: (props) => h('div', { className: 'components-base-control' },
      props.label ? h('label', { className: 'components-base-control__label' }, props.label) : null,
      h('input', {
        type: props.type || 'text',
        className: 'components-text-control__input',
        value: props.value != null ? props.value : '',
        min: props.min,
        max: props.max,
        step: props.step,
        placeholder: props.placeholder,
        onChange: (e) => props.onChange && props.onChange(e.target.value),
      })
    ),
    ToggleControl: (props) => h('label', { className: 'components-toggle-control' },
      h('input', {
        type: 'checkbox',
        checked: !!props.checked,
        onChange: (e) => props.onChange && props.onChange(e.target.checked),
      }),
      h('span', null, props.label || '')
    ),
    Button: (props) => h('button', {
      type: props.type || 'button',
      className: 'button' + (props.variant === 'primary' ? ' button-primary' : props.variant === 'secondary' ? ' button-secondary' : ''),
      disabled: props.disabled,
      onClick: (e) => props.onClick && props.onClick(e),
    }, props.children),
    Notice: (props) => h('div', { className: 'notice ' + (props.status === 'error' ? 'notice-error' : 'notice-success') }, h('p', null, props.children)),
    Spinner: () => h('span', { className: 'spinner is-active' }),
    Modal: (props) => h('div', { className: 'mytheme-fallback-modal' },
      h('div', { className: 'mytheme-fallback-modal__header' }, props.title || ''),
      h('div', { className: 'mytheme-fallback-modal__body' }, props.children),
      h('div', { className: 'mytheme-fallback-modal__footer' },
        h('button', { className: 'button', onClick: props.onRequestClose }, __('关闭', 'tanzanite-settings'))
      )
    ),
    SelectControl: (props) => h('div', { className: 'components-base-control' },
      props.label ? h('label', { className: 'components-base-control__label' }, props.label) : null,
      h('select', {
        className: 'components-select-control__input',
        value: props.value != null ? props.value : '',
        onChange: (e) => props.onChange && props.onChange(e.target.value),
        disabled: props.disabled,
      },
        (props.options || []).map((option) => h('option', {
          key: option.value,
          value: option.value,
        }, option.label))
      ),
      props.help ? h('p', { className: 'components-base-control__help' }, props.help) : null
    ),
    TextareaControl: (props) => h('div', { className: 'components-base-control' },
      props.label ? h('label', { className: 'components-base-control__label' }, props.label) : null,
      h('textarea', {
        className: 'components-textarea-control__input',
        value: props.value != null ? props.value : '',
        onChange: (e) => props.onChange && props.onChange(e.target.value),
        rows: props.rows || 4,
        placeholder: props.placeholder,
      }),
      props.help ? h('p', { className: 'components-base-control__help' }, props.help) : null
    ),
  }

  const TextControl = components.TextControl || Fallback.TextControl
  const ToggleControl = components.ToggleControl || Fallback.ToggleControl
  const Button = components.Button || Fallback.Button
  const Notice = components.Notice || Fallback.Notice
  const Spinner = components.Spinner || Fallback.Spinner
  const ModalComponent = components.Modal || Fallback.Modal
  const SelectControl = components.SelectControl || Fallback.SelectControl
  const TextareaControl = components.TextareaControl || Fallback.TextareaControl
  const { addQueryArgs } = wp.url
  const apiFetch = wp.apiFetch

  if (!window.TzLoyaltyConfig) {
    console.error('[Tanzanite] TzLoyaltyConfig not found')
    return
  }

  if (
    apiFetch &&
    typeof apiFetch.use === 'function' &&
    typeof apiFetch.createNonceMiddleware === 'function' &&
    TzLoyaltyConfig.nonce
  ) {
    apiFetch.use(apiFetch.createNonceMiddleware(TzLoyaltyConfig.nonce))
  }

  const REST_NAMESPACE = 'tanzanite/v1'
  const buildRestPath = (endpoint) => {
    const trimmed = String(endpoint || '').replace(/^\/+/, '')
    return `/${REST_NAMESPACE}/${trimmed}`
  }

  const ensureNonceHeader = (request) => {
    if (TzLoyaltyConfig.nonce) {
      request.headers = {
        ...(request.headers || {}),
        'X-WP-Nonce': TzLoyaltyConfig.nonce,
      }
    }
    return request
  }

  const buildQueryRouteUrl = (route) => {
    const rawSettings = TzLoyaltyConfig.settings || {}
    const root = (window.wpApiSettings && typeof window.wpApiSettings.root === 'string')
      ? window.wpApiSettings.root
      : ''
    const siteRoot = root ? root.replace(/\/wp-json\/?$/, '') : ''
    const base = siteRoot ? siteRoot.replace(/\/?$/, '') : ''
    const encodedRoute = encodeURIComponent(route)
    return base ? `${base}/?rest_route=${encodedRoute}` : `/?rest_route=${encodedRoute}`
  }

  const restRequest = async (endpoint, options = {}) => {
    if (!apiFetch) {
      return Promise.reject(new Error(__('无法连接到 WordPress API。', 'tanzanite-settings')))
    }

    const restRoute = buildRestPath(endpoint)
    const request = ensureNonceHeader({
      ...options,
      path: restRoute,
    })

    try {
      return await apiFetch(request)
    } catch (err) {
      if (err && err.code === 'invalid_json') {
        const fallbackUrl = buildQueryRouteUrl(restRoute)
        const retryRequest = ensureNonceHeader({
          ...options,
          url: fallbackUrl,
        })
        return apiFetch(retryRequest)
      }
      throw err
    }
  }

  const tierOrder = [
    'ordinary',
    'bronze',
    'silver',
    'gold',
    'platinum',
  ]

  const defaultRedeem = {
    enabled: false,
    percent_of_total: 5,
    value_per_point_base: 0.01,
    min_points: 0,
    stack_with_percent: true,
  }

  const tierDefaults = {
    ordinary: {
      label: __('Ordinary', 'tanzanite-settings'),
      name: __('Ordinary', 'tanzanite-settings'),
      min: 0,
      max: 499,
      discount: 0,
      products: [],
      categories: [],
      redeem: { ...defaultRedeem },
    },
    bronze: {
      label: __('Bronze', 'tanzanite-settings'),
      name: __('Bronze', 'tanzanite-settings'),
      min: 500,
      max: 1999,
      discount: 5,
      products: [],
      categories: [],
      redeem: { ...defaultRedeem },
    },
    silver: {
      label: __('Silver', 'tanzanite-settings'),
      name: __('Silver', 'tanzanite-settings'),
      min: 2000,
      max: 4999,
      discount: 10,
      products: [],
      categories: [],
      redeem: { ...defaultRedeem },
    },
    gold: {
      label: __('Gold', 'tanzanite-settings'),
      name: __('Gold', 'tanzanite-settings'),
      min: 5000,
      max: 9999,
      discount: 15,
      products: [],
      categories: [],
      redeem: { ...defaultRedeem },
    },
    platinum: {
      label: __('Platinum', 'tanzanite-settings'),
      name: __('Platinum', 'tanzanite-settings'),
      min: 10000,
      max: null,
      discount: 20,
      products: [],
      categories: [],
      redeem: { ...defaultRedeem },
    },
  }

  function normalizeItems(items) {
    if (!Array.isArray(items)) {
      return []
    }
    return items
      .map((item) => {
        if (item && typeof item === 'object') {
          const id = parseInt(item.id, 10)
          if (!id) {
            return null
          }
          return {
            id,
            name: String(item.name || item.title || ''),
            slug: item.slug ? String(item.slug) : undefined,
          }
        }
        const id = parseInt(item, 10)
        if (!id) {
          return null
        }
        return { id, name: '' }
      })
      .filter(Boolean)
  }

  function normalizeConfig(rawConfig) {
    const raw = rawConfig && typeof rawConfig === 'object' ? rawConfig : {}
    const settings = {
      enabled: raw.enabled !== false,
      apply_cart_discount: raw.apply_cart_discount !== false,
      points_per_unit: Number(raw.points_per_unit || 1),
      daily_checkin_points: Number(raw.daily_checkin_points != null ? raw.daily_checkin_points : 0),
      referral: {
        enabled: !!(raw.referral && raw.referral.enabled),
        bonus_inviter: Number(raw.referral && raw.referral.bonus_inviter != null ? raw.referral.bonus_inviter : 50),
        bonus_invitee: Number(raw.referral && raw.referral.bonus_invitee != null ? raw.referral.bonus_invitee : 30),
        token_ttl_days: Number(raw.referral && raw.referral.token_ttl_days != null ? raw.referral.token_ttl_days : 7),
        token_max_uses: Number(raw.referral && raw.referral.token_max_uses != null ? raw.referral.token_max_uses : 50),
      },
      tiers: {},
    }

    tierOrder.forEach((key) => {
      const incoming = raw.tiers && raw.tiers[key] ? raw.tiers[key] : {}
      const base = tierDefaults[key]
      const incomingRedeem = incoming.redeem && typeof incoming.redeem === 'object' ? incoming.redeem : {}
      const baseRedeem = base.redeem && typeof base.redeem === 'object' ? base.redeem : defaultRedeem
      settings.tiers[key] = {
        label: incoming.label || base.label,
        name: incoming.name || incoming.label || base.label,
        min: Number(incoming.min != null ? incoming.min : base.min),
        max: Number(incoming.max != null ? incoming.max : base.max),
        discount: Number(incoming.discount != null ? incoming.discount : base.discount),
        products: normalizeItems(incoming.products),
        categories: normalizeItems(incoming.categories),
        redeem: {
          enabled: incomingRedeem.enabled != null ? !!incomingRedeem.enabled : !!baseRedeem.enabled,
          percent_of_total: Number(incomingRedeem.percent_of_total != null ? incomingRedeem.percent_of_total : (baseRedeem.percent_of_total != null ? baseRedeem.percent_of_total : defaultRedeem.percent_of_total)),
          value_per_point_base: Number(incomingRedeem.value_per_point_base != null ? incomingRedeem.value_per_point_base : (baseRedeem.value_per_point_base != null ? baseRedeem.value_per_point_base : defaultRedeem.value_per_point_base)),
          min_points: Number(incomingRedeem.min_points != null ? incomingRedeem.min_points : (baseRedeem.min_points != null ? baseRedeem.min_points : defaultRedeem.min_points)),
          stack_with_percent: incomingRedeem.stack_with_percent != null ? !!incomingRedeem.stack_with_percent : (baseRedeem.stack_with_percent != null ? !!baseRedeem.stack_with_percent : !!defaultRedeem.stack_with_percent),
        },
      }
    })

    return settings
  }

  function serializeConfig(config) {
    const output = {
      enabled: !!config.enabled,
      apply_cart_discount: !!config.apply_cart_discount,
      points_per_unit: Number(config.points_per_unit || 0),
      daily_checkin_points: Number(config.daily_checkin_points != null ? config.daily_checkin_points : 0),
      referral: {
        enabled: !!(config.referral && config.referral.enabled),
        bonus_inviter: Number(config.referral && config.referral.bonus_inviter != null ? config.referral.bonus_inviter : 50),
        bonus_invitee: Number(config.referral && config.referral.bonus_invitee != null ? config.referral.bonus_invitee : 30),
        token_ttl_days: Number(config.referral && config.referral.token_ttl_days != null ? config.referral.token_ttl_days : 7),
        token_max_uses: Number(config.referral && config.referral.token_max_uses != null ? config.referral.token_max_uses : 50),
      },
      tiers: {},
    }

    tierOrder.forEach((key) => {
      const tier = config.tiers[key] || tierDefaults[key]
      const baseRedeem = tierDefaults[key] && tierDefaults[key].redeem ? tierDefaults[key].redeem : defaultRedeem
      const redeem = tier && tier.redeem && typeof tier.redeem === 'object' ? tier.redeem : baseRedeem
      output.tiers[key] = {
        label: tier.label,
        name: tier.name,
        min: Number(tier.min),
        max: Number(tier.max),
        discount: Number(tier.discount),
        products: (tier.products || []).map((item) => ({
          id: Number(item.id),
          name: String(item.name || ''),
          slug: item.slug ? String(item.slug) : undefined,
        })),
        categories: (tier.categories || []).map((item) => ({
          id: Number(item.id),
          name: String(item.name || ''),
          slug: item.slug ? String(item.slug) : undefined,
        })),
        redeem: {
          enabled: !!redeem.enabled,
          percent_of_total: Number(redeem.percent_of_total != null ? redeem.percent_of_total : (baseRedeem.percent_of_total != null ? baseRedeem.percent_of_total : defaultRedeem.percent_of_total)),
          value_per_point_base: Number(redeem.value_per_point_base != null ? redeem.value_per_point_base : (baseRedeem.value_per_point_base != null ? baseRedeem.value_per_point_base : defaultRedeem.value_per_point_base)),
          min_points: Number(redeem.min_points != null ? redeem.min_points : (baseRedeem.min_points != null ? baseRedeem.min_points : defaultRedeem.min_points)),
          stack_with_percent: redeem.stack_with_percent != null ? !!redeem.stack_with_percent : (baseRedeem.stack_with_percent != null ? !!baseRedeem.stack_with_percent : !!defaultRedeem.stack_with_percent),
        },
      }
    })

    return output
  }

  const NumberInput = (props) => h(TextControl, { ...props, type: 'number' })

  function ProductTokens({ items, onRemove }) {
    if (!items.length) {
      return h('p', { className: 'mytheme-loyalty-empty' }, __('未限制（默认全部适用）', 'tanzanite-settings'))
    }

    return h(
      'div',
      { className: 'mytheme-loyalty-token-container' },
      items.map((item) =>
        h('span', {
          key: `${item.id}`,
          className: 'mytheme-loyalty-token',
        },
          h('span', { className: 'mytheme-loyalty-token-label' }, item.name ? `${item.name}` : `#${item.id}`),
          h(Button, {
            isSmall: true,
            isLink: true,
            className: 'mytheme-loyalty-token-remove',
            onClick: () => onRemove(item),
          }, __('移除', 'tanzanite-settings'))
        )
      )
    )
  }

  function SearchModal({ type, tierKey, onClose, onSelect, initialItems }) {
    if (!ModalComponent) {
      return null
    }
    const [query, setQuery] = useState('')
    const [results, setResults] = useState([])
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState(null)

    const existingIds = useMemo(() => new Set((initialItems || []).map((item) => Number(item.id))), [initialItems])

    useEffect(() => {
      if (!query) {
        setResults([])
        return
      }
      let active = true
      async function run() {
        try {
          setLoading(true)
          setError(null)
          const path = type === 'product'
            ? addQueryArgs('/wp/v2/product', { search: query, per_page: 10 })
            : addQueryArgs('/wp/v2/product_cat', { search: query, per_page: 10 })
          const data = await apiFetch({ path })
          if (!active) {
            return
          }
          const mapped = (Array.isArray(data) ? data : []).map((item) => ({
            id: item.id,
            name: item.name || (item.title && item.title.rendered) || item.slug || item.id,
            slug: item.slug,
          }))
          setResults(mapped)
        } catch (err) {
          if (!active) {
            return
          }
          setError(err && err.message ? err.message : __('无法加载数据，请稍后再试。', 'tanzanite-settings'))
        } finally {
          if (active) {
            setLoading(false)
          }
        }
      }
      run()
      return () => {
        active = false
      }
    }, [query, type])

    const title = type === 'product'
      ? __('选择适用商品', 'tanzanite-settings')
      : __('选择适用商品分类', 'tanzanite-settings')

    return h(
      ModalComponent,
      { title, onRequestClose: onClose },
      h('div', { className: 'mytheme-loyalty-modal' },
        h(TextControl, {
          label: __('搜索', 'tanzanite-settings'),
          value: query,
          onChange: setQuery,
          placeholder: type === 'product'
            ? __('输入商品名称或 SKU', 'tanzanite-settings')
            : __('输入分类名称或别名', 'tanzanite-settings'),
        }),
        loading && h(Spinner, null),
        error && h(Notice, { status: 'error', onRemove: () => setError(null), isDismissible: true }, error),
        !loading && !error &&
          h('div', { className: 'mytheme-loyalty-search-results' },
            results.length === 0
              ? h('p', null, __('未找到结果', 'tanzanite-settings'))
              : results.map((item) =>
                h('div', {
                  key: item.id,
                  className: 'mytheme-loyalty-search-row',
                },
                  h('div', { className: 'mytheme-loyalty-search-info' },
                    h('strong', null, item.name || `#${item.id}`),
                    item.slug ? h('div', { className: 'mytheme-loyalty-slug' }, item.slug) : null
                  ),
                  h('div', { className: 'mytheme-loyalty-search-action' },
                    h(Button, {
                      variant: existingIds.has(Number(item.id)) ? 'tertiary' : 'primary',
                      disabled: existingIds.has(Number(item.id)),
                      onClick: () => {
                        if (!existingIds.has(Number(item.id))) {
                          onSelect(item)
                        }
                      },
                    }, existingIds.has(Number(item.id))
                      ? __('已添加', 'tanzanite-settings')
                      : __('添加', 'tanzanite-settings'))
                  )
                )
              )
          )
      )
    )
  }

  function TierCard({ tierKey, tier, onChange, onOpenModal, onRemoveProduct, onRemoveCategory }) {
    const header = tier.label || tier.name
    return h('section', { className: 'mytheme-loyalty-card' },
      h('h3', { className: 'mytheme-loyalty-card__title' }, header),
      h('div', { className: 'mytheme-loyalty-card__body' },
        h('div', { className: 'mytheme-loyalty-grid' },
          h(NumberInput, {
            label: __('最低积分', 'tanzanite-settings'),
            value: tier.min,
            onChange: (value) => onChange({ ...tier, min: Number(value) }),
            min: 0,
          }),
          h(NumberInput, {
            label: __('最高积分（-1 代表无限）', 'tanzanite-settings'),
            value: tier.max,
            onChange: (value) => onChange({ ...tier, max: Number(value) }),
          }),
          h(NumberInput, {
            label: __('折扣（%）', 'tanzanite-settings'),
            value: tier.discount,
            onChange: (value) => onChange({ ...tier, discount: Number(value) }),
            min: 0,
            max: 100,
          })
        ),

        // Redeem settings
        h('div', { className: 'mytheme-loyalty-scope' },
          h('h4', null, __('积分抵扣（按折扣前总额的百分比，且与等级折扣叠加）', 'tanzanite-settings')),
          h('div', { className: 'mytheme-loyalty-redeem-header' },
            h(ToggleControl, {
              label: __('启用该等级积分抵扣', 'tanzanite-settings'),
              checked: !!(tier.redeem && tier.redeem.enabled),
              onChange: (v) => onChange({ ...tier, redeem: { ...(tier.redeem||{}), enabled: !!v } }),
            }),
            h('span', { className: 'mytheme-loyalty-redeem-note' }, __('抵扣上限按折扣前总额计算，并与等级百分比折扣叠加。', 'tanzanite-settings'))
          ),
          h('div', { className: 'mytheme-loyalty-inline-grid mytheme-loyalty-inline-grid--redeem' },
            h(NumberInput, {
              label: __('抵扣上限百分比（基于折扣前总额）', 'tanzanite-settings'),
              value: tier.redeem && tier.redeem.percent_of_total != null ? tier.redeem.percent_of_total : 5,
              onChange: (value) => onChange({ ...tier, redeem: { ...(tier.redeem||{}), percent_of_total: Number(value) } }),
              min: 0,
              max: 100,
            }),
            h(NumberInput, {
              label: __('每积分面值（基准货币）', 'tanzanite-settings'),
              value: tier.redeem && tier.redeem.value_per_point_base != null ? tier.redeem.value_per_point_base : 0.01,
              onChange: (value) => onChange({ ...tier, redeem: { ...(tier.redeem||{}), value_per_point_base: Number(value) } }),
              min: 0,
              step: 0.0001,
            }),
            h(NumberInput, {
              label: __('最低可用积分', 'tanzanite-settings'),
              value: tier.redeem && tier.redeem.min_points != null ? tier.redeem.min_points : 0,
              onChange: (value) => onChange({ ...tier, redeem: { ...(tier.redeem||{}), min_points: Number(value) } }),
              min: 0,
            })
          ),
        ),

        h('div', { className: 'mytheme-loyalty-scope' },
          h('h4', null, __('适用商品', 'tanzanite-settings')),
          h(ProductTokens, {
            items: tier.products,
            onRemove: onRemoveProduct,
          }),
          h(Button, {
            variant: 'secondary',
            onClick: () => onOpenModal('product', tierKey),
          }, tier.products.length ? __('添加更多商品', 'tanzanite-settings') : __('指定适用商品…', 'tanzanite-settings'))
        ),
        h('div', { className: 'mytheme-loyalty-scope' },
          h('h4', null, __('适用商品分类', 'tanzanite-settings')),
          h(ProductTokens, {
            items: tier.categories,
            onRemove: onRemoveCategory,
          }),
          h(Button, {
            variant: 'secondary',
            onClick: () => onOpenModal('category', tierKey),
          }, tier.categories.length ? __('添加更多分类', 'tanzanite-settings') : __('指定适用分类…', 'tanzanite-settings'))
        )
      )
    )
  }

  function App() {
    const [config, setConfig] = useState(normalizeConfig(TzLoyaltyConfig.settings))
    const [searchModal, setSearchModal] = useState(null)
    const [notice, setNotice] = useState(null)
    useEffect(() => {
      const input = document.getElementById('tz-loyalty-config')
      if (input) {
        input.value = JSON.stringify(serializeConfig(config))
      }
    }, [config])

    useEffect(() => {
      const form = document.getElementById('tz-loyalty-form')
      if (!form) {
        return
      }
      const handleSubmit = (event) => {
        const serialized = serializeConfig(config)
        const input = document.getElementById('tz-loyalty-config')
        if (input) {
          input.value = JSON.stringify(serialized)
        }
      }
      form.addEventListener('submit', handleSubmit)
      return () => form.removeEventListener('submit', handleSubmit)
    }, [config])

    const setTier = (key, nextTier) => {
      setConfig((prev) => ({
        ...prev,
        tiers: {
          ...prev.tiers,
          [key]: {
            ...prev.tiers[key],
            ...nextTier,
          },
        },
      }))
    }

    const removeItem = (tierKey, type, target) => {
      setConfig((prev) => ({
        ...prev,
        tiers: {
          ...prev.tiers,
          [tierKey]: {
            ...prev.tiers[tierKey],
            [type]: prev.tiers[tierKey][type].filter((item) => Number(item.id) !== Number(target.id)),
          },
        },
      }))
    }

    const addItem = (tierKey, type, item) => {
      setConfig((prev) => {
        const tier = prev.tiers[tierKey]
        const list = tier[type]
        if (list.some((entry) => Number(entry.id) === Number(item.id))) {
          return prev
        }
        return {
          ...prev,
          tiers: {
            ...prev.tiers,
            [tierKey]: {
              ...tier,
              [type]: [...list, { id: Number(item.id), name: item.name || '', slug: item.slug || undefined }],
            },
          },
        }
      })
      setNotice({
        status: 'success',
        message: type === 'product'
          ? __('已添加商品到该等级。', 'tanzanite-settings')
          : __('已添加分类到该等级。', 'tanzanite-settings'),
      })
    }

    const saveSettings = async () => {
      setNotice(null)
      try {
        const serialized = serializeConfig(config)
        await restRequest('loyalty/config', {
          method: 'POST',
          data: serialized,
        })
        setNotice({ status: 'success', message: __('已保存设置。', 'tanzanite-settings') })
      } catch (err) {
        setNotice({ status: 'error', message: err && err.message ? err.message : __('保存失败，请稍后再试。', 'tanzanite-settings') })
      }
    }

    return h(Fragment, null,
      notice && h(Notice, {
        status: notice.status,
        isDismissible: true,
        onRemove: () => setNotice(null),
      }, notice.message),
      h('div', { className: 'mytheme-loyalty-global' },
        h('div', { className: 'mytheme-loyalty-global__item' },
          h(ToggleControl, {
            label: __('启用积分功能', 'tanzanite-settings'),
            checked: config.enabled,
            onChange: (value) => setConfig((prev) => ({ ...prev, enabled: value })),
          })
        ),
        h('div', { className: 'mytheme-loyalty-global__item' },
          h(ToggleControl, {
            label: __('在购物车自动应用折扣', 'tanzanite-settings'),
            checked: config.apply_cart_discount,
            onChange: (value) => setConfig((prev) => ({ ...prev, apply_cart_discount: value })),
            help: __('若关闭，则仍可通过 API/模板手动使用折扣信息。', 'tanzanite-settings'),
          })
        ),
        h('div', { className: 'mytheme-loyalty-global__item' },
          h(NumberInput, {
            label: __('每消费 1 货币单位积多少积分', 'tanzanite-settings'),
            value: config.points_per_unit,
            min: 0,
            step: 0.1,
            onChange: (value) => setConfig((prev) => ({ ...prev, points_per_unit: Number(value) })),
          })
        ),
        h('div', { className: 'mytheme-loyalty-global__item' },
          h(NumberInput, {
            label: __('每日签到一次积累多少积分（只允许每日一次，签到积分30天自动清零）', 'tanzanite-settings'),
            value: config.daily_checkin_points != null ? config.daily_checkin_points : 0,
            min: 0,
            onChange: (value) => setConfig((prev) => ({ ...prev, daily_checkin_points: Math.max(0, Number(value)) })),
          })
        )
      ,
        h('div', { className: 'mytheme-loyalty-scope' },
          h('h4', null, __('邀请注册设置', 'tanzanite-settings')),
          h(ToggleControl, {
            label: __('启用邀请注册奖励', 'tanzanite-settings'),
            checked: !!(config.referral && config.referral.enabled),
            onChange: (v) => setConfig((prev) => ({ ...prev, referral: { ...(prev.referral||{}), enabled: !!v } })),
          }),
          h('div', { className: 'mytheme-loyalty-inline-grid' },
            h(NumberInput, {
              label: __('邀请人奖励积分', 'tanzanite-settings'),
              value: config.referral && config.referral.bonus_inviter != null ? config.referral.bonus_inviter : 50,
              min: 0,
              onChange: (value) => setConfig((prev) => ({ ...prev, referral: { ...(prev.referral||{}), bonus_inviter: Math.max(0, Number(value)) } })),
            }),
            h(NumberInput, {
              label: __('被邀请人奖励积分', 'tanzanite-settings'),
              value: config.referral && config.referral.bonus_invitee != null ? config.referral.bonus_invitee : 30,
              min: 0,
              onChange: (value) => setConfig((prev) => ({ ...prev, referral: { ...(prev.referral||{}), bonus_invitee: Math.max(0, Number(value)) } })),
            }),
            h(NumberInput, {
              label: __('邀请链接有效期（天）', 'tanzanite-settings'),
              value: config.referral && config.referral.token_ttl_days != null ? config.referral.token_ttl_days : 7,
              min: 1,
              onChange: (value) => setConfig((prev) => ({ ...prev, referral: { ...(prev.referral||{}), token_ttl_days: Math.max(1, Number(value)) } })),
            }),
            h(NumberInput, {
              label: __('每个邀请链接最大使用次数', 'tanzanite-settings'),
              value: config.referral && config.referral.token_max_uses != null ? config.referral.token_max_uses : 50,
              min: 1,
              onChange: (value) => setConfig((prev) => ({ ...prev, referral: { ...(prev.referral||{}), token_max_uses: Math.max(1, Number(value)) } })),
            })
          ),
          h('p', { className: 'description' }, __('邀请链接将指向 /register，并在用户注册成功后自动发放邀请人与被邀请人奖励。', 'tanzanite-settings'))
        )
      ),
      tierOrder.map((key) =>
        h(TierCard, {
          key,
          tierKey: key,
          tier: config.tiers[key],
          onChange: (nextTier) => setTier(key, nextTier),
          onOpenModal: (type) => setSearchModal({ type, tierKey: key }),
          onRemoveProduct: (item) => removeItem(key, 'products', item),
          onRemoveCategory: (item) => removeItem(key, 'categories', item),
        })
      ),
      searchModal && h(SearchModal, {
        type: searchModal.type === 'product' ? 'product' : 'category',
        tierKey: searchModal.tierKey,
        onClose: () => setSearchModal(null),
        onSelect: (item) => {
          addItem(searchModal.tierKey, searchModal.type === 'product' ? 'products' : 'categories', item)
        },
        initialItems: config.tiers[searchModal.tierKey][searchModal.type === 'product' ? 'products' : 'categories'],
      })
    )
  }

  document.addEventListener('DOMContentLoaded', () => {
    const target = document.getElementById('tz-loyalty-app')
    if (!target) {
      console.error('[Tanzanite] Loyalty app container not found')
      return
    }
    try {
      wp.element.render(h(App), target)
    } catch (err) {
      console.error('[Tanzanite] Loyalty admin UI failed to render:', err)
      target.innerHTML = '<div class="notice notice-error"><p>' +
        (__('无法渲染忠诚度设置界面，请刷新页面或清空缓存。', 'tanzanite-settings')) +
        '</p></div>'
    }
  })
})()
