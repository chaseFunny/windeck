import { REPO_URL, type Content } from '../lib/content';

// The footer, signed with the hz mark: the small solid bubble and "hz",
// the h in the text's color, the z in the violet.
export default function SiteFooter(props: { t: Content }) {
  return (
    <footer class="footer">
      <div class="wrap footer__inner">
        <span>{props.t.footer.tagline}</span>
        <span class="footer__links">
          <a href={`${REPO_URL}/blob/main/docs/data.md`}>{props.t.footer.privacy}</a>
          <a href={`${REPO_URL}/issues`}>{props.t.footer.contact}</a>
          <span>© 2026 Windeck</span>
          <span class="maker">
            {props.t.footer.made}
            <a class="hz" href="https://github.com/hz7002" target="_blank" rel="noopener" aria-label="hz on GitHub">
              <svg viewBox="0 0 40 32" width="20" height="16" aria-hidden="true">
                <path d="M12 0H28A12 12 0 0 1 40 12V20A12 12 0 0 1 28 32H3A3 3 0 0 1 0 29V12A12 12 0 0 1 12 0Z" fill="currentColor" />
              </svg>
              <b>
                h<i>z</i>
              </b>
            </a>
          </span>
        </span>
      </div>
    </footer>
  );
}
