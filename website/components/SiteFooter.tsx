import type { Content } from '../lib/content';

export default function SiteFooter(props: { t: Content }) {
  return (
    <footer class="footer">
      <div class="wrap footer__inner">
        <span>{props.t.footer.tagline}</span>
        <span class="footer__links">
          <a href="https://github.com/windeck/windeck/blob/main/docs/data.md">{props.t.footer.privacy}</a>
          <a href="mailto:hello@windeck.app">{props.t.footer.contact}</a>
          <span>© 2026 Windeck</span>
        </span>
      </div>
    </footer>
  );
}
