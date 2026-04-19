import { expect, test } from "@playwright/test";

test("article list flow covers search, detail, and csv export", async ({ page }) => {
  await page.goto("/");

  await expect(page.getByRole("heading", { name: "技術情報の一覧・検索" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Platform Weekly Digest" })).toBeVisible();

  await page.getByPlaceholder("タイトル・概要で検索").fill("platform");
  await page.getByPlaceholder("category 例: kubernetes").fill("kubernetes");
  await page.getByRole("combobox").selectOption({ label: "E2E Tech Blog" });
  await page.getByLabel("未読のみ").check();
  await page.getByLabel("お気に入りのみ").check();
  await page.getByRole("button", { name: "Search" }).click();

  // list と export が同じ filter 条件を使う前提なので、一覧でも 1 件に絞れることを先に確認する。
  await expect(page.getByRole("link", { name: "Platform Weekly Digest" })).toBeVisible();
  await expect(page.getByRole("link", { name: "SRE Incident Review" })).toHaveCount(0);

  await page.getByRole("link", { name: "Platform Weekly Digest" }).click();
  await expect(page).toHaveURL(/\/articles\/1001$/);
  await expect(page.getByRole("heading", { name: "Platform Weekly Digest" })).toBeVisible();

  await page.goto("/");
  await page.getByPlaceholder("タイトル・概要で検索").fill("platform");
  await page.getByPlaceholder("category 例: kubernetes").fill("kubernetes");
  await page.getByRole("combobox").selectOption({ label: "E2E Tech Blog" });
  await page.getByLabel("未読のみ").check();
  await page.getByLabel("お気に入りのみ").check();
  await page.getByRole("button", { name: "Search" }).click();

  const downloadPromise = page.waitForEvent("download");
  await page.getByRole("link", { name: "CSV Download" }).click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toMatch(/^articles-\d{8}-\d{6}\.csv$/);

  const csvPath = await download.path();
  expect(csvPath).not.toBeNull();
  const csv = await download.createReadStream();
  expect(csv).not.toBeNull();

  const chunks: Buffer[] = [];
  for await (const chunk of csv!) {
    chunks.push(Buffer.from(chunk));
  }
  const body = Buffer.concat(chunks).toString("utf8");
  expect(body).toContain("title,url,source_name,category,published_at,fetched_at,is_read,is_favorite,excerpt,tags");
  expect(body).toContain("Platform Weekly Digest");
});
