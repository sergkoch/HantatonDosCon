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
});
