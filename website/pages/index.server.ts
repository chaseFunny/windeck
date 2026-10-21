import { defineHead } from 'void';

export const head = defineHead(() => ({
  title: 'Your windows, like browser tabs',
  meta: [
    {
      name: 'description',
      content:
        'Windeck is a window switcher for macOS: every window as a readable tab grouped by app, a large preview to be sure, and a quick tap back to the previous window.',
    },
    { property: 'og:image', content: '/icon.png' },
  ],
}));
