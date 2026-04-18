import { api, buildArticleExportUrl, buildArticleQueryParams } from "./api";

describe("api helpers", () => {
  beforeEach(() => {
    api.defaults.baseURL = "http://localhost:8080/";
  });

  it("omits empty values and normalizes filters", () => {
    const from = "2026-04-16T09:30";
    const to = "2026-04-17T18:45";

    expect(
      buildArticleQueryParams({
        q: "kubernetes",
        category: "",
        source_id: 4,
        is_read: false,
        is_favorite: true,
        from,
        to,
      }),
    ).toEqual({
      q: "kubernetes",
      category: undefined,
      source_id: 4,
      is_read: false,
      is_favorite: true,
      from: new Date(from).toISOString(),
      to: new Date(to).toISOString(),
      sort: undefined,
      order: undefined,
      page: undefined,
    });
  });

  it("builds the CSV export URL from normalized query params", () => {
    const from = "2026-04-16T09:30";

    expect(
      buildArticleExportUrl({
        q: "platform",
        source_id: 7,
        is_read: false,
        from,
      }),
    ).toBe(
      `http://localhost:8080/api/v1/exports/articles.csv?q=platform&source_id=7&is_read=false&from=${encodeURIComponent(
        new Date(from).toISOString(),
      )}`,
    );
  });
});
