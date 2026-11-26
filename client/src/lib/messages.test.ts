import { describe, expect, it } from "vitest";
import { translate } from "./messages";

describe("translate", () => {
  it("returns Russian strings by default dictionary", () => {
    expect(translate("ru", "layout.navCheck")).toBe("Проверить");
  });

  it("returns English strings", () => {
    expect(translate("en", "layout.navCheck")).toBe("Check");
  });

  it("interpolates parameters", () => {
    expect(translate("ru", "report.added", { count: 3 })).toBe("3 новых");
    expect(translate("en", "report.added", { count: 3 })).toBe("3 new");
  });

  it("keeps plural forms readable in both languages", () => {
    expect(translate("ru", "report.critical", { count: 1 })).toBe("1 критично");
    expect(translate("en", "report.critical", { count: 1 })).toBe("1 critical");
  });

  it("translates cookie banner and contact labels", () => {
    expect(translate("ru", "cookie.accept")).toBe("Принять");
    expect(translate("en", "cookie.accept")).toBe("Accept");
    expect(translate("ru", "layout.contacts")).toBe("Контакты команды:");
    expect(translate("en", "layout.contacts")).toBe("Team contacts:");
  });

  it("translates API error codes", () => {
    expect(translate("ru", "error.invalidInput")).toBe("Проверьте адрес и глубину");
    expect(translate("en", "error.invalidInput")).toBe("Check the address and depth");
    expect(translate("en", "error.http", { status: 413 })).toBe("Error 413");
  });
});
