import { defineHead } from 'void';

export const head = defineHead(() => ({
  title: '窗签 — 像浏览器标签页一样切换窗口',
  htmlAttrs: { lang: 'zh-CN' },
  meta: [
    {
      name: 'description',
      content: '窗签（Windeck）是一款 macOS 窗口切换器：按应用分组的完整标题标签、中央高清大预览、轻按一下回到上一个窗口。',
    },
    { property: 'og:image', content: '/icon.png' },
  ],
}));
