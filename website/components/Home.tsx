import { For } from 'solid-js';
import { DOWNLOAD_URL, content, type Lang } from '../lib/content';
import Icon, { icons } from './Icon';
import SwitcherDemo from './SwitcherDemo';
import SiteHeader from './SiteHeader';
import SiteFooter from './SiteFooter';

// The page, as the hz landing pages go: the nav, the first screen with its
// one big bubble and the product on a gray block, three cards, text and
// pictures taking turns, facts, the price, questions, the dark ending and
// the footer. Its one action is downloading.

const featureIcons = [icons.list, icons.image, icons.layers, icons.pin, icons.keyboard, icons.feather];

export default function Home(props: { lang: Lang }) {
  const t = content[props.lang];
  return (
    <div class="site" lang={props.lang === 'zh' ? 'zh-CN' : 'en'}>
      <SiteHeader t={t} />
      <main>
        <section class="hero wrap">
          <p class="eyebrow">{t.hero.eyebrow}</p>
          <h1 class="bubble display">{t.hero.title}</h1>
          <p class="lead">{t.hero.lead}</p>
          <div class="actions">
            <a class="button button--primary button--large" href={DOWNLOAD_URL}>
              <Icon d={icons.download} size={20} />
              {t.hero.download}
            </a>
            <a class="button button--large" href="#how">
              {t.hero.how}
            </a>
          </div>
          <p class="note">{t.hero.note}</p>
        </section>
        <div class="wrap">
          <div class="showcase">
            <SwitcherDemo t={t} />
          </div>
        </div>

        <section id="how" class="wrap section">
          <h2 class="h2">{t.stepsTitle}</h2>
          <div class="cards">
            <For each={t.steps}>
              {(s) => (
                <article class="card">
                  <div class="card__keys">
                    <For each={s.keys}>{(k) => <kbd class="key">{k}</kbd>}</For>
                  </div>
                  <h3 class="h3">{s.title}</h3>
                  <p>{s.body}</p>
                </article>
              )}
            </For>
          </div>
        </section>

        <section id="features" class="wrap section split">
          <div class="split__text">
            <h2 class="h2">{t.featuresTitle}</h2>
            <p class="section-lead">{t.featuresLead}</p>
            <ul class="features">
              <For each={t.features}>
                {(f, i) => (
                  <li>
                    <span class="features__icon">
                      <Icon d={featureIcons[i()]} size={20} />
                    </span>
                    <div>
                      <strong>{f.title}</strong>
                      <p>{f.body}</p>
                    </div>
                  </li>
                )}
              </For>
            </ul>
          </div>
          <div class="split__art" aria-hidden="true">
            <div class="badges">
              <For
                each={[
                  { title: t.demo.minimized, icon: icons.minimized },
                  { title: t.demo.desktop, icon: icons.desktop },
                  { title: props.lang === 'zh' ? '外接显示器' : 'External display', icon: icons.monitor },
                  { title: props.lang === 'zh' ? '已隐藏' : 'Hidden', icon: icons.eyeOff },
                ]}
              >
                {(b, i) => (
                  <div class="badges__row" classList={{ 'is-sel': i() === 1 }}>
                    <span class="badges__app" />
                    <span class="badges__line" style={{ width: `${[62, 48, 70, 54][i()]}%` }} />
                    <span class="tag tag--gray">
                      <Icon d={b.icon} size={14} />
                      {b.title}
                    </span>
                  </div>
                )}
              </For>
            </div>
          </div>
        </section>

        <section id="smart" class="wrap section split split--flip">
          <div class="split__text">
            <h2 class="h2">{t.smart.title}</h2>
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
          <div class="split__art" aria-hidden="true">
            <div class="cycle">
              <For each={t.smart.cycle}>
                {(c, i) => (
                  <div class="cycle__item" classList={{ 'is-first': i() === 0 }}>
                    <span class="cycle__num">{i() + 1}</span>
                    <span class="cycle__label">{c}</span>
                    {i() === 0 && <kbd class="key key--small">⇥</kbd>}
                    {i() === 1 && (
                      <span class="cycle__star">
                        <Icon d={icons.star} size={16} fill />
                      </span>
                    )}
                  </div>
                )}
              </For>
            </div>
          </div>
        </section>

        <section id="privacy" class="wrap section split">
          <div class="split__text">
            <h2 class="h2">{t.privacy.title}</h2>
            <p class="section-lead">{t.privacy.lead}</p>
            <ul class="checks">
              <For each={t.privacy.points}>
                {(p) => (
                  <li>
                    <Icon d={icons.check} size={20} />
                    <span>{p}</span>
                  </li>
                )}
              </For>
            </ul>
          </div>
          <div class="split__art" aria-hidden="true">
            <div class="privacy-list">
              <p class="privacy-list__title">
                <Icon d={icons.shieldCheck} size={20} />
                {t.privacy.listTitle}
              </p>
              <For each={t.privacy.list}>
                {(name) => (
                  <div class="privacy-list__row">
                    <span class="badges__app" />
                    <span>{name}</span>
                    <span class="tag tag--gray">
                      <Icon d={icons.eyeOff} size={14} />
                      {t.privacy.never}
                    </span>
                  </div>
                )}
              </For>
            </div>
          </div>
        </section>

        <section class="wrap section">
          <h2 class="h2">{t.facts.title}</h2>
          <div class="facts">
            <For each={t.facts.items}>
              {(s) => (
                <div class="fact">
                  <strong>{s.value}</strong>
                  <span>{s.label}</span>
                </div>
              )}
            </For>
          </div>
        </section>

        <section id="pricing" class="wrap section">
          <h2 class="h2">{t.pricing.title}</h2>
          <p class="section-lead">{t.pricing.lead}</p>
          <div class="plans">
            <For each={[t.pricing.free, t.pricing.pro]}>
              {(p, i) => (
                <div class="plan">
                  <div class="plan__head">
                    <h3 class="h3">{p.name}</h3>
                    {i() === 1 && <span class="tag">14 {props.lang === 'zh' ? '天试用' : '-day trial'}</span>}
                  </div>
                  <p class="plan__price">
                    {p.price} <span>{p.note}</span>
                  </p>
                  <ul class="checks checks--small">
                    <For each={p.items}>
                      {(it) => (
                        <li>
                          <Icon d={icons.check} size={16} />
                          <span>{it}</span>
                        </li>
                      )}
                    </For>
                  </ul>
                  {i() === 1 ? (
                    <span class="button button--disabled" aria-disabled="true">
                      {t.pricing.buy}
                    </span>
                  ) : (
                    <a class="button" href={DOWNLOAD_URL}>
                      {t.nav.download}
                    </a>
                  )}
                </div>
              )}
            </For>
          </div>
          <p class="note">{t.pricing.early}</p>
        </section>

        <section id="faq" class="wrap section faq">
          <h2 class="h2">{t.faqTitle}</h2>
          <For each={t.faq}>
            {(f) => (
              <details>
                <summary>
                  {f.q}
                  <Icon d={icons.chevronDown} size={20} />
                </summary>
                <p>{f.a}</p>
              </details>
            )}
          </For>
        </section>

        <section class="ending">
          <div class="wrap ending__inner">
            <img src="/icon.png" width="72" height="72" alt="" />
            <h2 class="h2">{t.cta.title}</h2>
            <p>{t.cta.lead}</p>
            <a class="button button--primary button--large" href={DOWNLOAD_URL}>
              <Icon d={icons.download} size={20} />
              {t.cta.button}
            </a>
          </div>
        </section>
      </main>
      <SiteFooter t={t} />
    </div>
  );
}
