import { For } from 'solid-js';
import { BUY_URL, DOWNLOAD_URL, content, type Lang } from '../lib/content';
import SwitcherDemo from './SwitcherDemo';
import SiteHeader from './SiteHeader';
import SiteFooter from './SiteFooter';

const featureIcons = [
  'M3 5h10M3 8h7M3 11h9',
  'M2.5 3.5h11v9h-11zM2.5 6h11',
  'M3 3h4v4H3zM9 3h4v4H9zM3 9h4v4H3zM9 9h4v4H9z',
  'M8 2v12M2 8h12',
  'M4 4l8 8M12 4l-8 8',
  'M9 2 4 9h4l-1 5 5-7H8z',
];

export default function Home(props: { lang: Lang }) {
  const t = content[props.lang];
  return (
    <div class="site" lang={props.lang === 'zh' ? 'zh-CN' : 'en'}>
      <SiteHeader t={t} />
      <main>
        <section class="hero wrap">
          <p class="eyebrow">{t.hero.eyebrow}</p>
          <h1 class="display">
            {t.hero.title[0]}
            <br />
            <span class="muted">{t.hero.title[1]}</span>
          </h1>
          <p class="lead">{t.hero.lead}</p>
          <div class="actions">
            <a class="button button--primary" href={DOWNLOAD_URL}>
              <svg viewBox="0 0 16 16" width="15" height="15" fill="currentColor" aria-hidden="true">
                <path d="M11.2 8.5c0-1.6 1.3-2.4 1.4-2.4-.8-1.1-2-1.3-2.4-1.3-1-.1-2 .6-2.5.6s-1.3-.6-2.2-.6C4.4 4.8 3 5.8 3 7.9c0 1.3.5 2.7 1.1 3.6.5.8 1 1.5 1.7 1.5.7 0 1-.4 1.8-.4s1.1.4 1.8.4c.8 0 1.2-.7 1.7-1.5.5-.7.7-1.4.7-1.5 0 0-1.6-.6-1.6-2.5zM9.6 3.6c.4-.5.7-1.1.6-1.8-.6 0-1.3.4-1.7.9-.4.4-.7 1.1-.6 1.7.7.1 1.3-.3 1.7-.8z" />
              </svg>
              {t.hero.download}
            </a>
            <code class="brew">{t.hero.brew}</code>
          </div>
          <p class="note">{t.hero.note}</p>
          <SwitcherDemo t={t} />
        </section>

        <section class="wrap steps">
          <For each={t.steps}>
            {(s) => (
              <div class="step">
                <div class="step__keys">
                  <For each={s.keys}>{(k) => <kbd class="key key--small">{k}</kbd>}</For>
                </div>
                <h3>{s.title}</h3>
                <p>{s.body}</p>
              </div>
            )}
          </For>
        </section>

        <section id="features" class="wrap section">
          <h2 class="title">{t.featuresTitle}</h2>
          <p class="section-lead">{t.featuresLead}</p>
          <div class="grid">
            <For each={t.features}>
              {(f, i) => (
                <article class="card">
                  <span class="card__icon">
                    <svg viewBox="0 0 16 16" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                      <path d={featureIcons[i()]} />
                    </svg>
                  </span>
                  <h3>{f.title}</h3>
                  <p>{f.body}</p>
                </article>
              )}
            </For>
          </div>
        </section>

        <section id="smart" class="wrap section split">
          <div>
            <h2 class="title">{t.smart.title}</h2>
            <p class="section-lead">{t.smart.lead}</p>
            <ul class="points">
              <For each={t.smart.points}>
                {(p) => (
                  <li>
                    <strong>{p.title}</strong>
                    <span>{p.body}</span>
                  </li>
                )}
              </For>
            </ul>
          </div>
          <div class="cycle" aria-hidden="true">
            <For each={t.smart.cycle}>
              {(c, i) => (
                <div class="cycle__item" classList={{ 'is-first': i() === 0, 'is-pred': i() === 1 }}>
                  <span class="cycle__num">{i() + 1}</span>
                  <span>{c}</span>
                  {i() === 0 && <span class="cycle__lock">⇥</span>}
                  {i() === 1 && <span class="cycle__star">★</span>}
                </div>
              )}
            </For>
          </div>
        </section>

        <section id="privacy" class="wrap section privacy">
          <h2 class="title">{t.privacy.title}</h2>
          <p class="section-lead">{t.privacy.lead}</p>
          <ul class="checks">
            <For each={t.privacy.points}>{(p) => <li>{p}</li>}</For>
          </ul>
        </section>

        <section class="wrap section">
          <h2 class="title center">{t.perf.title}</h2>
          <div class="stats">
            <For each={t.perf.items}>
              {(s) => (
                <div class="stat">
                  <strong>{s.value}</strong>
                  <span>{s.label}</span>
                </div>
              )}
            </For>
          </div>
        </section>

        <section id="pricing" class="wrap section">
          <h2 class="title center">{t.pricing.title}</h2>
          <p class="section-lead center">{t.pricing.lead}</p>
          <div class="plans">
            <For each={[t.pricing.free, t.pricing.pro]}>
              {(p, i) => (
                <div class="plan" classList={{ 'plan--pro': i() === 1 }}>
                  <h3>{p.name}</h3>
                  <p class="plan__price">
                    {p.price} <span>{p.note}</span>
                  </p>
                  <ul class="checks">
                    <For each={p.items}>{(it) => <li>{it}</li>}</For>
                  </ul>
                  {i() === 1 ? (
                    <a class="button button--primary" href={BUY_URL}>
                      {t.pricing.buy}
                    </a>
                  ) : (
                    <a class="button" href={DOWNLOAD_URL}>
                      {t.nav.download}
                    </a>
                  )}
                </div>
              )}
            </For>
          </div>
          <p class="note center">{t.pricing.early}</p>
        </section>

        <section id="faq" class="wrap section faq">
          <h2 class="title center">FAQ</h2>
          <For each={t.faq}>
            {(f) => (
              <details>
                <summary>{f.q}</summary>
                <p>{f.a}</p>
              </details>
            )}
          </For>
        </section>

        <section class="wrap cta">
          <img src="/icon.png" width="88" height="88" alt="" />
          <h2 class="title">{t.cta.title}</h2>
          <a class="button button--primary" href={DOWNLOAD_URL}>
            {t.cta.button}
          </a>
        </section>
      </main>
      <SiteFooter t={t} />
    </div>
  );
}
