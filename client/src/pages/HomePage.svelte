<script lang="ts">
  import { onMount } from "svelte";
  import { createScan } from "../api";
  import { navigate } from "../lib/router.svelte";
  import { t } from "../lib/i18n.svelte";
  import Button from "../components/Button.svelte";
  import Card from "../components/Card.svelte";

  let url = $state("");
  let depth = $state(1);
  let error = $state("");
  let busy = $state(false);
  let urlInput: HTMLInputElement;

  onMount(() => {
    if (window.location.hash !== "#how") return;

    const section = document.getElementById("how");
    if (!section) return;

    const reduce = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    section.scrollIntoView({
      behavior: reduce ? "auto" : "smooth",
      block: "start",
    });
    section.focus({ preventScroll: true });
  });

  async function onSubmit(event: SubmitEvent) {
    event.preventDefault();
    error = "";
    busy = true;

    try {
      const scan = await createScan(url, depth);
      navigate(`/report/${scan.id}`);
    } catch (cause) {
      error =
        cause instanceof Error ? cause.message : t("home.error");
      urlInput?.focus();
    } finally {
      busy = false;
    }
  }
</script>

<div class="container">
  <section class="hero">
    <div class="hero__badge">
      <span class="hero__badge-dot" aria-hidden="true"></span>
      {t("home.badge")}
    </div>

    <h1 class="hero__title">
      {t("home.titleBefore")}<span class="hero__title-accent"
        >{t("home.titleAccent")}</span
      >{t("home.titleAfter")}
    </h1>
    <p class="hero__lead">
      {t("home.lead")}
    </p>

    <form
      class="form"
      onsubmit={onSubmit}
      aria-describedby={error ? "scan-hint scan-error" : "scan-hint"}
      aria-busy={busy}
      {...{
        toolname: "scan_website_accessibility",
        tooldescription:
          "Запускает проверку доступности сайта по URL и глубине обхода",
      }}
    >
      <div class="field">
        <label class="field__label" for="url">{t("home.urlLabel")}</label>
        <input
          class="input"
          id="url"
          name="url"
          type="url"
          inputmode="url"
          placeholder="https://example.ru"
          bind:value={url}
          bind:this={urlInput}
          aria-invalid={error ? "true" : undefined}
          {...{
            toolparamdescription:
              "Полный адрес сайта, начиная с http:// или https://",
          }}
          required
        />
      </div>

      <div class="field field--small">
        <label class="field__label" for="depth">{t("home.depthLabel")}</label>
        <select
          class="select"
          id="depth"
          name="depth"
          bind:value={depth}
          {...{
            toolparamdescription:
              "Глубина обхода: 1 — одна страница, 2 — ссылки, 3 — глубже",
          }}
        >
          <option value={1}>{t("home.depth1")}</option>
          <option value={2}>{t("home.depth2")}</option>
          <option value={3}>{t("home.depth3")}</option>
        </select>
      </div>

      <Button type="submit" disabled={busy}>
        {busy ? t("home.busy") : t("home.submit")}
      </Button>

      <p class="field__hint form__hint" id="scan-hint">
        {t("home.hint")}
      </p>
    </form>

    {#if error}
      <p class="field__error" id="scan-error" role="alert">{error}</p>
    {/if}
  </section>

  <section class="section" id="how" tabindex="-1" aria-labelledby="how-title">
    <div class="how">
      <div class="how__head">
        <div>
          <p class="how__eyebrow">{t("home.howEyebrow")}</p>
          <h2 class="section__title" id="how-title">{t("home.howTitle")}</h2>
        </div>
      </div>

      <ol class="how__steps">
        <li class="how-step">
          <span class="how-step__num" aria-hidden="true"
            ><span class="how-step__digit">1</span></span
          >
          <h3 class="how-step__title">{t("home.step1Title")}</h3>
          <p class="how-step__text">
            {t("home.step1Text")}
          </p>
        </li>
        <li class="how-step">
          <span class="how-step__num" aria-hidden="true"
            ><span class="how-step__digit">2</span></span
          >
          <h3 class="how-step__title">{t("home.step2Title")}</h3>
          <p class="how-step__text">
            {t("home.step2Text")}
          </p>
        </li>
        <li class="how-step">
          <span class="how-step__num" aria-hidden="true"
            ><span class="how-step__digit">3</span></span
          >
          <h3 class="how-step__title">{t("home.step3Title")}</h3>
          <p class="how-step__text">
            {t("home.step3Text")}
          </p>
        </li>
      </ol>
    </div>
  </section>

  <section class="section" id="checks" aria-labelledby="checks-title">
    <h2 class="section__title" id="checks-title">{t("home.checksTitle")}</h2>
    <div class="grid">
      <Card title={t("home.cardImages")}>
        {#snippet icon()}
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
            <rect
              x="3"
              y="3"
              width="18"
              height="18"
              rx="3"
              stroke="currentColor"
              stroke-width="1.8"
            />
            <circle cx="8.5" cy="8.5" r="1.6" fill="currentColor" />
            <path
              d="m21 15-5-5L5 21"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        {/snippet}
        {t("home.cardImagesLead")}<code class="code">alt</code>{t(
          "home.cardImagesTail",
        )}
      </Card>

      <Card title={t("home.cardPdf")}>
        {#snippet icon()}
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
            <path
              d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linejoin="round"
            />
            <path
              d="M14 3v5h5"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linejoin="round"
            />
            <path
              d="M8.5 13h7M8.5 17h4"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
            />
          </svg>
        {/snippet}
        {t("home.cardPdfText")}
      </Card>

      <Card title={t("home.cardLinks")}>
        {#snippet icon()}
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
            <path
              d="M10 13a5 5 0 0 0 7.07 0l2.83-2.83a5 5 0 0 0-7.07-7.07l-1.41 1.41"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
            <path
              d="M14 11a5 5 0 0 0-7.07 0L4.1 13.83a5 5 0 0 0 7.07 7.07l1.41-1.41"
              stroke="currentColor"
              stroke-width="1.8"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        {/snippet}
        {t("home.cardLinksText")}
      </Card>
    </div>
  </section>
</div>
