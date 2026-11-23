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
      navigate(`/scan/${scan.id}`);
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

    <form class="form" onsubmit={onSubmit} aria-describedby="scan-hint">
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
          required
        />
      </div>

      <div class="field field--small">
        <label class="field__label" for="depth">Глубина</label>
        <select class="select" id="depth" name="depth" bind:value={depth}>
          <option value={1}>1 — только страница</option>
          <option value={2}>2 — страница и ссылки</option>
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
    <h2 class="section__title" id="how-title">Что проверяем</h2>
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
