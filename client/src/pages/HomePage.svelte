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
    <p class="hero__eyebrow">Проверка доступности</p>
    <h1 class="hero__title">Сайт, который понимают все</h1>
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
        <p class="field__hint" id="scan-hint">
          Только http или https. Начните с главной страницы.
        </p>
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
    </form>

    {#if error}
      <p class="field__error" role="alert">{error}</p>
    {/if}
  </section>

  <section class="section" id="how" aria-labelledby="how-title">
    <h2 class="section__title" id="how-title">Что проверяем</h2>
    <div class="grid">
      <Card title="Изображения">
        Есть ли <code class="code">alt</code>, осмыслен ли он и совпадает ли с
        картинкой.
      </Card>
      <Card title="PDF">
        Есть ли текстовый слой, теги структуры и альтернативная версия.
      </Card>
      <Card title="Ссылки">Работают ли они и понятно ли, куда ведут.</Card>
    </div>
  </section>
</div>
