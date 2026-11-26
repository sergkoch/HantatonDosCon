import { mount } from "svelte";
import App from "./App.svelte";
import "./styles.css";

const target = document.getElementById("app");

if (!target) {
  throw new Error("Root element not found");
}

// Убираем статический каркас из index.html перед монтированием Svelte.
target.replaceChildren();

const app = mount(App, { target });

export default app;
