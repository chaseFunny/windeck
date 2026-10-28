import { DOWNLOAD_URL, REPO_URL, SITE_URL, type Lang } from './content';

// What search engines and link previews read of a page: its canonical
// address and its twin in the other language, the share card, and the
// app described as structured data (schema.org SoftwareApplication).

const words = {
  en: {
    title: 'Windeck — switch windows on macOS like browser tabs',
    description:
      'Windeck is a window switcher for macOS: hold ⌥ and press Tab to see every window as a readable tab grouped by app, with a large preview. Minimized, hidden and other-desktop windows included. Free.',
    keywords: 'macOS window switcher, alt tab for mac, cmd tab alternative, switch windows mac, window manager mac, Windeck',
    locale: 'en_US',
    path: '/',
  },
  zh: {
    title: '窗签 Windeck — 像切浏览器标签页一样切换 Mac 窗口',
    description:
      '窗签（Windeck）是一款 macOS 窗口切换器：按住 ⌥ 再按 Tab，所有窗口按应用分组、以完整标题的标签呈现，中央大预览帮你确认。支持最小化、隐藏和其他桌面的窗口，免费下载。',
    keywords: 'Mac 窗口切换, macOS 窗口切换器, Alt Tab Mac, Cmd Tab 替代, Mac 多窗口管理, 窗签, Windeck',
    locale: 'zh_CN',
    path: '/zh',
  },
};

export function seoHead(lang: Lang) {
  const w = words[lang];
  const url = SITE_URL + w.path;
  const image = `${SITE_URL}/og.png`;
  const app = {
    '@context': 'https://schema.org',
    '@type': 'SoftwareApplication',
    name: 'Windeck',
    alternateName: '窗签',
    applicationCategory: 'UtilitiesApplication',
    operatingSystem: 'macOS 13 or later',
    softwareVersion: '1.2.0',
    description: w.description,
    url,
    downloadUrl: DOWNLOAD_URL,
    image,
    sameAs: [REPO_URL],
    inLanguage: ['en', 'zh-CN'],
    offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
  };
  return {
    title: w.title,
    htmlAttrs: { lang: lang === 'zh' ? 'zh-CN' : 'en' },
    meta: [
      { name: 'description', content: w.description },
      { name: 'keywords', content: w.keywords },
      { property: 'og:type', content: 'website' },
      { property: 'og:site_name', content: 'Windeck' },
      { property: 'og:title', content: w.title },
      { property: 'og:description', content: w.description },
      { property: 'og:url', content: url },
      { property: 'og:image', content: image },
      { property: 'og:image:width', content: '1200' },
      { property: 'og:image:height', content: '630' },
      { property: 'og:locale', content: w.locale },
      { name: 'twitter:card', content: 'summary_large_image' },
      { name: 'twitter:title', content: w.title },
      { name: 'twitter:description', content: w.description },
      { name: 'twitter:image', content: image },
    ],
    link: [
      { rel: 'canonical', href: url },
      { rel: 'alternate', hreflang: 'en', href: `${SITE_URL}/` },
      { rel: 'alternate', hreflang: 'zh-CN', href: `${SITE_URL}/zh` },
      { rel: 'alternate', hreflang: 'x-default', href: `${SITE_URL}/` },
    ],
    script: [{ type: 'application/ld+json', innerHTML: JSON.stringify(app) }],
  };
}
