<script lang="ts">
  import { createScan } from "../api";
  import { navigate } from "../lib/router.svelte";
  import Button from "../components/Button.svelte";
  import Card from "../components/Card.svelte";

  let url = $state("");
  let depth = $state(1);
  let error = $state("");
  let busy = $state(false);

  async function onSubmit(event: SubmitEvent) {
    event.preventDefault();
    error = "";
    busy = true;

    try {
      const scan = await createScan(url, depth);
      navigate(`/report/${scan.id}`);
    } catch (cause) {
      error =
        cause instanceof Error ? cause.message : "Не удалось запустить проверку";
    } finally {
      busy = false;
    }
  }
</script>

<div class="container">
  <section class="hero">
    <div class="hero__badge">
      <span class="hero__badge-dot" aria-hidden="true"></span>
      Проверка доступности · WCAG
    </div>

    <h1 class="hero__title">
      Сайт, который <span class="hero__title-accent">понимают</span> все
    </h1>
    <p class="hero__lead">
      Проверьте изображения, PDF и ссылки. Найдите не только формальные ошибки, но
      и смысловые: описание не совпадает с картинкой, ссылка ведёт не туда.
    </p>

    <form
      class="form"
      onsubmit={onSubmit}
      aria-describedby="scan-hint"
      {...{
        toolname: "scan_website_accessibility",
        tooldescription:
          "Запускает проверку доступности сайта по URL и глубине обхода",
      }}
    >
      <div class="field">
        <label class="field__label" for="url">Адрес сайта</label>
        <input
          class="input"
          id="url"
          name="url"
          type="url"
          inputmode="url"
          placeholder="https://example.ru"
          bind:value={url}
          {...{
            toolparamdescription:
              "Полный адрес сайта, начиная с http:// или https://",
          }}
          required
        />
      </div>

      <div class="field field--small">
        <label class="field__label" for="depth">Глубина</label>
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
          <option value={1}>1 — страница</option>
          <option value={2}>2 — ссылки</option>
          <option value={3}>3 — глубже</option>
        </select>
      </div>

      <Button type="submit" disabled={busy}>
        {busy ? "Запускаем…" : "Проверить"}
      </Button>

      <p class="field__hint form__hint" id="scan-hint">
        Только http или https. Начните с главной страницы.
      </p>
    </form>

    {#if error}
      <p class="field__error" role="alert">{error}</p>
    {/if}
  </section>

  <section class="section" id="how" aria-labelledby="how-title">
    <div class="how">
      <div class="how__head">
        <div>
          <p class="how__eyebrow">Процесс</p>
          <h2 class="section__title" id="how-title">Как это работает</h2>
        </div>
        <p class="how__note">3 шага · без настройки</p>
      </div>

      <ol class="how__steps">
        <li class="how-step">
          <span class="how-step__num" aria-hidden="true"
            ><span class="how-step__digit">1</span></span
          >
          <h3 class="how-step__title">Вставьте адрес</h3>
          <p class="how-step__text">
            Вставьте адрес сайта. Можно проверить одну страницу или пройти по
            ссылкам с неё.
          </p>
        </li>
        <li class="how-step">
          <span class="how-step__num" aria-hidden="true"
            ><span class="how-step__digit">2</span></span
          >
          <h3 class="how-step__title">Мы проверим сайт</h3>
          <p class="how-step__text">
            Соберём страницы, картинки, PDF и ссылки. Смотрим только этот сайт,
            чтобы не создавать ему лишнюю нагрузку.
          </p>
        </li>
        <li class="how-step">
          <span class="how-step__num" aria-hidden="true"
            ><span class="how-step__digit">3</span></span
          >
          <h3 class="how-step__title">Получите отчёт</h3>
          <p class="how-step__text">
            Покажем, что не так, насколько это важно и как исправить. Отчёт
            можно скачать или оставить в нём только нужные пункты.
          </p>
        </li>
      </ol>
    </div>
  </section>

  <section class="section" id="checks" aria-labelledby="checks-title">
    <h2 class="section__title" id="checks-title">Что проверяем</h2>
    <div class="grid">
      <Card title="Изображения">
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
        Есть ли <code class="code">alt</code>, осмыслен ли он и совпадает ли с
        картинкой.
      </Card>

      <Card title="PDF">
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
        Есть ли текстовый слой, теги структуры и альтернативная версия.
      </Card>

      <Card title="Ссылки">
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
        Работают ли они и понятно ли, куда ведут.
      </Card>
    </div>
  </section>
</div>
