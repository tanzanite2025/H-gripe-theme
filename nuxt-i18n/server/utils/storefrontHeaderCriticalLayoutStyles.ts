/** Keep the header usable before its component and global styles arrive. */
export const storefrontHeaderCriticalLayoutStyles = `
:where(.site-header-root){position:fixed;inset:0 0 auto;z-index:900;width:100%;color:var(--tz-text-primary,#14202b);font-family:inherit}
:where(.site-header-root,.site-header-root *){box-sizing:border-box}
:where(.site-header-surface){width:100%;padding:0;background:var(--tz-input-surface,#f1f5f9)}
:where(.site-header-mainbar){display:grid;grid-template-columns:minmax(0,clamp(10rem,14vw,17.5rem)) minmax(0,1fr) minmax(0,clamp(10rem,14vw,17.5rem));align-items:center;gap:clamp(.5rem,1vw,1rem);height:var(--tz-site-header-desktop-mainbar-height,64px);min-height:var(--tz-site-header-desktop-mainbar-height,64px);padding:0 2rem;border-bottom:1px solid rgba(20,32,43,.12)}
:where(.site-header-mainbar>div:first-child){display:flex;align-items:center;min-width:0}
:where(.site-header-brand){display:inline-flex;align-items:center;min-width:0;max-width:100%;text-decoration:none;background:transparent}
:where(.site-header-brand--desktop){width:17.5rem;height:3.4rem}
:where(.site-header-brand__image){display:block;width:100%;height:100%;max-width:100%;max-height:100%;object-fit:contain;object-position:left center;background:transparent}
:where([data-header-mega-nav]){display:flex;align-items:center;justify-content:center;gap:clamp(.25rem,.9vw,1.25rem);min-width:0}
:where([data-header-mega-nav]>.hidden){display:none}
:where(.site-header-menu-laser){display:inline-flex;align-items:center;justify-content:center;gap:.375rem;min-width:0;padding:.5rem clamp(.5rem,.6vw,.75rem) .75rem;border:0;background:transparent;color:inherit;font:inherit;font-size:clamp(.875rem,1.15vw,1.125rem);font-weight:800;line-height:1;cursor:pointer}
:where(.site-header-menu-laser__text){min-width:0;overflow:hidden;text-overflow:ellipsis;text-transform:uppercase;white-space:nowrap}
:where(.site-header-menu-laser__icon){width:.875rem;height:.875rem;flex-shrink:0}
:where(.site-header-actions){--site-header-action-width:4.5rem;--site-header-action-height:2.25rem;--site-header-action-icon-size:1.375rem;display:flex;align-items:center;justify-content:flex-end;gap:.38rem;min-width:0}
:where(.site-header-action-cell){position:relative;display:inline-flex;align-items:center;justify-content:center;flex:0 0 var(--site-header-action-width,4.5rem);width:var(--site-header-action-width,4.5rem);height:var(--site-header-action-height,2.25rem);min-width:0}
:where(.site-header-action-button){display:inline-flex;align-items:center;justify-content:center;gap:.34rem;width:100%;height:100%;padding:0 .5rem;border:0;border-radius:9999px;background:transparent;color:inherit;font:inherit;line-height:1;cursor:pointer}
:where(.site-header-search-trigger__icon,.site-header-language-trigger svg,.site-header-language-trigger img){display:block;width:var(--site-header-action-icon-size,1.375rem);height:var(--site-header-action-icon-size,1.375rem);flex-shrink:0}
:where(.site-header-language-trigger>span:first-child){display:inline-flex;align-items:center;justify-content:center;flex-shrink:0}
:where(.site-header-search-hint){position:absolute;top:calc(100% + .55rem);left:50%;z-index:1250;display:grid;gap:.25rem;width:max-content;max-width:min(18rem,72vw);padding:.72rem .86rem;transform:translateX(-50%);border:1px solid rgba(20,32,43,.14);border-radius:10px;background:#fff;opacity:0;pointer-events:none}
:where(.site-header-action-cell--search:hover .site-header-search-hint,.site-header-action-cell--search:focus-within .site-header-search-hint){opacity:1}
:where(.site-header-search-hint__title){font-size:.78rem;font-weight:720;line-height:1.25}
:where(.site-header-search-hint__body){font-size:.72rem;line-height:1.35}
:where(.site-header-breadcrumb-row){display:flex;align-items:center;width:100%;height:var(--tz-site-header-desktop-breadcrumb-height,31px);min-width:0;background:var(--tz-input-surface,#f1f5f9)}
:where(.site-header-breadcrumb-list){display:flex;align-items:center;gap:.375rem;width:100%;min-width:0;margin:0;padding:0 2rem;list-style:none;font-size:.875rem;line-height:1.25}
:where(.site-header-breadcrumb-item,.breadcrumb-subnav-trigger){display:inline-flex;align-items:center;gap:.25rem;min-width:0}
:where(.site-header-breadcrumb-item svg){width:1rem;height:1rem;flex-shrink:0}
:where(.breadcrumb-subnav-trigger){padding:.25rem;border:0;background:transparent;color:inherit;font:inherit;cursor:pointer}
:where(.site-header-layout-spacer){width:100%;height:var(--tz-site-header-spacer-height,95px);flex:0 0 auto;pointer-events:none}
:where(.site-header-mobile-surface){width:100%;height:var(--tz-site-header-mobile-topbar-height,50px);margin:0;padding:0}
:where(.site-header-mobile-topbar){display:grid;grid-template-columns:repeat(3,minmax(0,1fr));align-items:center;width:100%;height:var(--tz-site-header-mobile-topbar-height,50px);padding:0}
:where(.site-header-mobile-shop-slot,.site-header-mobile-brand-slot){display:flex;align-items:center;min-width:0}
:where(.site-header-mobile-shop-slot){gap:clamp(.25rem,2vw,.5rem)}
:where(.site-header-mobile-brand-slot,.site-header-brand--mobile){justify-content:center}
:where(.site-header-brand--mobile){width:100%;height:48px}
:where(.site-header-brand--mobile .site-header-brand__image){width:48px;height:48px;max-width:48px;max-height:48px;object-position:center}
:where(.site-header-mobile-shop-trigger,.site-header-mobile-tools-trigger,.site-header-mobile-guides-trigger){display:inline-flex;align-items:center;justify-content:center;width:2.25rem;height:2.25rem;padding:0;border:0;background:transparent;color:inherit;cursor:pointer}
:where(.site-header-mobile-shop-trigger svg,.site-header-mobile-tools-trigger svg,.site-header-mobile-guides-trigger svg){width:1.25rem;height:1.25rem}
:where(.site-header-actions--mobile){gap:clamp(.875rem,4.5vw,1.25rem);width:100%;padding-inline-end:max(.25rem,env(safe-area-inset-right))}
:where(.site-header-actions--mobile .site-header-action-cell){flex-basis:2.25rem;width:2.25rem}
:where(.site-header-mobile-nav-row){width:100%;height:var(--tz-site-header-mobile-nav-height,44px);margin:0;padding:0;background:var(--tz-mobile-bottom-chrome-surface,#f1f5f9)}
:where(.site-header-mobile-nav){display:grid;grid-template-columns:none;grid-auto-columns:minmax(0,1fr);grid-auto-flow:column;width:100%;height:100%;padding:0}
:where(.site-header-mobile-nav__button){display:inline-flex;align-items:center;justify-content:center;gap:.18rem;min-width:0;height:100%;padding:0 .125rem;overflow:hidden;border:0;background:transparent;color:inherit;font:inherit;font-size:.8125rem;font-weight:600;text-transform:uppercase;line-height:1;cursor:pointer}
:where(.site-header-mobile-nav__button>span){min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
:where(.site-header-mobile-nav__button svg){width:.875rem;height:.875rem;flex-shrink:0}
:where(.site-header-language-label--mobile){display:none}
.site-header-root .site-header-surface{padding:0}
.site-header-root .site-header-mobile-surface{margin:0;padding:0}
.site-header-root .site-header-mobile-nav-row{margin:0;padding:0}
@media(max-width:1279px){
:where(.site-header-mainbar){grid-template-columns:minmax(0,1fr) minmax(0,1fr);grid-template-rows:50px 44px;height:94px;gap:0;padding:0 .75rem}
:where(.site-header-mainbar>[data-header-mega-nav]){grid-column:1/-1;grid-row:2;justify-content:space-around;gap:0}
:where(.site-header-mainbar>.site-header-actions){grid-column:2;grid-row:1}
:where(.site-header-brand--desktop){width:min(100%,10.5rem);height:48px}
:where(.site-header-menu-laser){padding:.5rem .125rem;font-size:clamp(.625rem,2vw,.875rem)}
:where(.site-header-breadcrumb-row){height:var(--tz-site-header-mobile-breadcrumb-height,32px)}
:where(.site-header-breadcrumb-list){padding:0 .75rem}
:where(.site-header-layout-spacer){height:var(--tz-site-header-spacer-height,126px)}
}
`
