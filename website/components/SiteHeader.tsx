import type { Content } from '../lib/content';
import { DOWNLOAD_URL } from '../lib/content';

// The nav: 64 high, the mark at the left, the one action at the right.
export default function SiteHeader(props: { t: Content }) {
  const n = () => props.t.nav;
  return (
    <header class="header">
      <div class="wrap header__inner">
        <a class="brand" href={n().otherHref === '/' ? '/zh' : '/'}>
          <img src="/icon.png" width="28" height="28" alt="" />
          <span>Windeck</span>
        </a>
        <nav class="nav" aria-label="Sections">
          <a href="#features">{n().features}</a>
          <a href="#smart">{n().smart}</a>
          <a href="#privacy">{n().privacy}</a>
          <a href="#pricing">{n().pricing}</a>
          <a href="#faq">{n().faq}</a>
        </nav>
        <div class="header__end">
          <a class="lang" href={n().otherHref}>
            {n().other}
          </a>
          <a class="button button--primary" href={DOWNLOAD_URL}>
            {n().download}
          </a>
        </div>
      </div>
    </header>
  );
}
