import { describe, expect, it } from "vitest";
import { kindLabel, scoreLabel } from "./score";

describe("scoreLabel", () => {
  it("returns a good label for a high score", () => {
    expect(scoreLabel(90)).toBe("Хорошо");
  });

  it("returns a warning label for a middle score", () => {
    expect(scoreLabel(60)).toBe("Нужно исправить");
  });

  it("returns a critical label for a low score", () => {
    expect(scoreLabel(20)).toBe("Критично");
  });

  it("handles a missing score", () => {
    expect(scoreLabel(null)).toBe("Проверка ещё идёт");
  });

  it("returns English labels", () => {
    expect(scoreLabel(90, "en")).toBe("Good");
    expect(scoreLabel(60, "en")).toBe("Needs work");
    expect(scoreLabel(20, "en")).toBe("Critical");
    expect(scoreLabel(null, "en")).toBe("Scan in progress");
  });
});

describe("kindLabel", () => {
  it("translates finding kinds", () => {
    expect(kindLabel("image")).toBe("Изображение");
    expect(kindLabel("pdf")).toBe("PDF");
    expect(kindLabel("link")).toBe("Ссылка");
  });

  it("translates finding kinds to English", () => {
    expect(kindLabel("image", "en")).toBe("Image");
    expect(kindLabel("page", "en")).toBe("Page");
    expect(kindLabel("table", "en")).toBe("Table");
  });
});
