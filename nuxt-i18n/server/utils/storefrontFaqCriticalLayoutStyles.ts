/** Small SSR fallback; component styles take over when their assets arrive. */
export const storefrontFaqCriticalLayoutStyles = `
:where(.page-faq,.home-faq,.faqs-page,.storefront-faq-placeholder){box-sizing:border-box;width:100%;min-width:0;color:var(--tz-text-primary,#14202b);font-family:inherit;overflow-wrap:anywhere}
:where(.page-faq,.faqs-page,.home-faq__surface,.storefront-faq-placeholder){background:var(--tz-card-surface,#fff);border-radius:1.25rem;padding:1.25rem}
:where(.page-faq,.storefront-faq-placeholder){border:1px solid rgba(20,32,43,.12)}
:where(.page-faq-slot){width:100%;margin-top:1.25rem}
:where(.page-faq__header,.home-faq__header,.faqs-page__header){margin-bottom:1.25rem}
:where(.page-faq__title,.home-faq__title,.faqs-page__title){margin:0;font-size:1.25rem;font-weight:800;line-height:1.3}
:where(.page-faq__subtitle,.home-faq__subtitle,.faqs-page__intro){margin:.5rem 0;color:var(--tz-text-secondary,#475569)}
.page-faq__desktop-list,.home-faq__desktop-layout,.global-all-faqs-desktop-grouped-search-results-panel{display:none!important}
:where(.page-faq__mobile-list,.home-faq__mobile-preview,.global-all-faqs-mobile-grouped-search-results-panel){display:block}
:where(.page-faq__question,.home-faq__question,.global-all-faqs-mobile-grouped-search-results-panel__item-button,.desktop-faq-master-detail__button){display:flex;align-items:center;justify-content:space-between;gap:.75rem;box-sizing:border-box;width:100%;min-width:0;padding:.9rem 1rem;border:0;background:transparent;color:inherit;font:inherit;text-align:left;cursor:pointer}
:where(.page-faq__item,.home-faq__item,.global-all-faqs-mobile-grouped-search-results-panel__item,.desktop-faq-master-detail__item){border:1px solid rgba(20,32,43,.12);border-radius:.75rem}
:where(.page-faq,.home-faq,.faqs-page) svg{width:1.1rem;height:1.1rem;flex-shrink:0}
:where(.page-faq,.home-faq,.faqs-page) img{max-width:100%;height:auto}
:where(.page-faq__question-text,.desktop-faq-master-detail__question,.global-all-faqs-mobile-grouped-search-results-panel__question){min-width:0;overflow-wrap:anywhere;line-height:1.45}
:where(.desktop-faq-master-detail){display:grid;grid-template-columns:minmax(0,1fr);gap:.75rem;min-width:0}
:where(.desktop-faq-master-detail__list,.home-faq__sidebar-tabs,.faqs-tabs){display:flex;flex-direction:column;gap:.55rem;min-width:0}
:where(.desktop-faq-master-detail__detail){box-sizing:border-box;min-width:0;min-height:14rem;padding:1rem;border:1px solid rgba(20,32,43,.12);border-radius:1rem}
:where(.desktop-faq-master-detail__detail-empty){display:flex;align-items:center;justify-content:center;gap:.65rem;min-height:12rem}
:where(.home-faq__sidebar-button,.faqs-tabs__button){box-sizing:border-box;width:100%;padding:.7rem;border:1px solid rgba(20,32,43,.12);border-radius:.75rem;background:transparent;color:inherit;text-align:left}
:where(.home-faq__group-title){display:none}
:where(.faqs-search__input){box-sizing:border-box;width:100%;padding:.75rem;margin-bottom:1rem;border:1px solid rgba(20,32,43,.15);border-radius:.75rem}
:where(.faqs-layout){display:grid;gap:1rem;min-width:0}
:where(.faqs-content,.home-faq__desktop-panel){min-width:0}
:where(.storefront-faq-placeholder){min-height:10rem}
:where(.storefront-faq-placeholder--with-saved-answers){min-height:0;margin-bottom:1rem}
:where(.storefront-faq-placeholder__message){margin:0 0 1rem;line-height:1.5}
:where(.storefront-faq-placeholder__layout,.storefront-faq-placeholder__questions,.storefront-faq-placeholder__question){display:grid;gap:.75rem}
:where(.storefront-faq-placeholder__question){padding:1rem;border:1px solid rgba(20,32,43,.1);border-radius:.75rem}
:where(.storefront-faq-placeholder__line){display:block;height:.65rem;border-radius:1rem;background:rgba(20,32,43,.14)}
:where(.storefront-faq-placeholder__line--short){width:55%}
:where(.storefront-faq-placeholder__answer){display:none}
:where(.storefront-faq-placeholder__retry){padding:.6rem 1rem;border:1px solid currentColor;border-radius:.75rem;background:transparent;color:var(--tz-text-accent,#047857);font:inherit;cursor:pointer}
@media(min-width:768px){.page-faq__mobile-list,.home-faq__mobile-preview,.global-all-faqs-mobile-grouped-search-results-panel{display:none!important}.page-faq__desktop-list,.global-all-faqs-desktop-grouped-search-results-panel{display:grid!important}.home-faq__desktop-layout{display:grid!important;grid-template-columns:minmax(12rem,.3fr) minmax(0,1fr);gap:1rem}:where(.storefront-faq-placeholder--loading,.page-faq__desktop-list){min-height:26rem}}
@media(min-width:981px){:where(.desktop-faq-master-detail){grid-template-columns:minmax(0,.9fr) minmax(0,1.1fr)}:where(.faqs-layout){grid-template-columns:clamp(15rem,21vw,18rem) minmax(0,1fr)}}
`
