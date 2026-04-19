import { expect, test } from "@playwright/test";

test("notification list flow covers read update", async ({ page }) => {
  await page.goto("/notifications");

  await expect(page.getByRole("heading", { name: "通知一覧" })).toBeVisible();
  const notification = page.locator("article").filter({ hasText: "Collector failed for E2E feed" });
  await expect(notification).toBeVisible();

  await notification.getByRole("button", { name: "既読にする" }).click();
  await expect(notification.getByText("既読")).toBeVisible();

  // mutation 後に一覧 query が invalidation される前提なので、read filter 側でも同じ通知が見えることを確認する。
  await page.getByLabel("Status").selectOption("read");
  await expect(page.locator("article").filter({ hasText: "Collector failed for E2E feed" })).toBeVisible();
});
