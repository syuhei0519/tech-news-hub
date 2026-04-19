import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { SourceManagementPage } from "./SourceManagementPage";
import { renderWithProviders } from "../test/render";
import { server } from "../test/setup";
import { buildSource } from "../test/fixtures";
import { sourcesHandler } from "../test/handlers";

describe("SourceManagementPage", () => {
  it("populates the edit form from the selected source and resets to create mode", async () => {
    const user = userEvent.setup();
    const sources = [buildSource()];

    server.use(sourcesHandler(() => sources));

    renderWithProviders(<SourceManagementPage />);

    expect(await screen.findByText("Kubernetes Blog")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Edit Source" }));

    expect(screen.getByRole("heading", { name: "Edit: Kubernetes Blog" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("Kubernetes Blog");
    expect(screen.getByLabelText("Fetch URL")).toHaveValue("https://example.com/kubernetes.xml");

    await user.click(screen.getByRole("button", { name: "Reset" }));
    expect(screen.getByRole("heading", { name: "Create Source" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("");

    await user.click(screen.getByRole("button", { name: "Edit Source" }));
    await user.click(screen.getByRole("button", { name: "New Source" }));
    expect(screen.getByRole("heading", { name: "Create Source" })).toBeInTheDocument();
    expect(screen.getByLabelText("Name")).toHaveValue("");
  });

  it("shows validation errors before submitting", async () => {
    const user = userEvent.setup();
    let createCalls = 0;

    server.use(
      sourcesHandler(() => []),
      http.post("http://localhost:8080/api/v1/sources", () => {
        createCalls += 1;
        return HttpResponse.json(buildSource({ id: 2 }));
      }),
    );

    renderWithProviders(<SourceManagementPage />);

    expect(await screen.findByText("0 sources")).toBeInTheDocument();
    await user.type(screen.getByLabelText("Name"), "   ");
    await user.type(screen.getByLabelText("Fetch URL"), "not-a-url");
    await user.type(screen.getByLabelText("Default Category"), "   ");
    await user.click(screen.getByRole("button", { name: "Create Source" }));

    // trim 後に空になる入力を使い、zod validation が submit を止めることを確認する。
    expect(await screen.findByText("name is required")).toBeInTheDocument();
    expect(screen.getByText("fetch_url must be a valid URL")).toBeInTheDocument();
    expect(screen.getByText("default_category is required")).toBeInTheDocument();
    expect(createCalls).toBe(0);
  });

  it("submits trimmed payloads for create and edit", async () => {
    const user = userEvent.setup();
    let sources = [buildSource()];
    let createdPayload: Record<string, unknown> | undefined;
    let updatedPayload: Record<string, unknown> | undefined;

    server.use(
      sourcesHandler(() => sources),
      http.post("http://localhost:8080/api/v1/sources", async ({ request }) => {
        createdPayload = (await request.json()) as Record<string, unknown>;
        const created = buildSource({
          id: 2,
          name: String(createdPayload.name),
          fetch_url: String(createdPayload.fetch_url),
          interval_minutes: Number(createdPayload.interval_minutes),
          default_category: String(createdPayload.default_category),
          is_enabled: Boolean(createdPayload.is_enabled),
        });
        sources = [...sources, created];
        return HttpResponse.json(created);
      }),
      http.patch("http://localhost:8080/api/v1/sources/:id", async ({ params, request }) => {
        updatedPayload = (await request.json()) as Record<string, unknown>;
        const id = Number(params.id);
        const updated = buildSource({
          id,
          name: String(updatedPayload.name),
          fetch_url: String(updatedPayload.fetch_url),
          interval_minutes: Number(updatedPayload.interval_minutes),
          default_category: String(updatedPayload.default_category),
          is_enabled: Boolean(updatedPayload.is_enabled),
        });
        sources = sources.map((source) => (source.id === id ? updated : source));
        return HttpResponse.json(updated);
      }),
    );

    renderWithProviders(<SourceManagementPage />);
    expect(await screen.findByText("Kubernetes Blog")).toBeInTheDocument();

    await user.clear(screen.getByLabelText("Name"));
    await user.type(screen.getByLabelText("Name"), "  New Feed  ");
    await user.clear(screen.getByLabelText("Fetch URL"));
    await user.type(screen.getByLabelText("Fetch URL"), "  https://example.com/new.xml  ");
    await user.clear(screen.getByLabelText("Default Category"));
    await user.type(screen.getByLabelText("Default Category"), "  platform  ");
    await user.click(screen.getByRole("button", { name: "Create Source" }));

    // create / edit どちらも form 値を trim して payload に乗せる現在仕様を固定する。
    await waitFor(() =>
      expect(createdPayload).toMatchObject({
        name: "New Feed",
        fetch_url: "https://example.com/new.xml",
        default_category: "platform",
      }),
    );
    expect(await screen.findByRole("heading", { name: "Edit: New Feed" })).toBeInTheDocument();

    const sourceCards = screen.getAllByText("Edit Source");
    await user.click(sourceCards[0]);
    await user.clear(screen.getByLabelText("Name"));
    await user.type(screen.getByLabelText("Name"), "  Kubernetes Blog Updated  ");
    await user.click(screen.getByRole("button", { name: "Save Changes" }));

    await waitFor(() =>
      expect(updatedPayload).toMatchObject({
        name: "Kubernetes Blog Updated",
      }),
    );
  });

  it("toggles source enable state with the full update payload", async () => {
    const user = userEvent.setup();
    let sources = [buildSource()];
    let togglePayload: Record<string, unknown> | undefined;

    server.use(
      sourcesHandler(() => sources),
      http.patch("http://localhost:8080/api/v1/sources/1", async ({ request }) => {
        togglePayload = (await request.json()) as Record<string, unknown>;
        sources = [buildSource({ is_enabled: false })];
        return HttpResponse.json(sources[0]);
      }),
    );

    renderWithProviders(<SourceManagementPage />);

    const sourceCard = await screen.findByText("Kubernetes Blog");
    await user.click(within(sourceCard.closest("article")!).getByRole("button", { name: "Enabled" }));

    // toggle 専用 API ではなく完全更新 API を使うため、他フィールドが保持された payload になる。
    await waitFor(() =>
      expect(togglePayload).toEqual({
        name: "Kubernetes Blog",
        type: "rss",
        fetch_url: "https://example.com/kubernetes.xml",
        fetch_method: "rss",
        interval_minutes: 60,
        default_category: "kubernetes",
        is_enabled: false,
      }),
    );
    expect(await screen.findByRole("button", { name: "Disabled" })).toBeInTheDocument();
  });

  it("shows list and toggle API errors", async () => {
    const user = userEvent.setup();

    server.use(
      sourcesHandler(() => [buildSource()]),
      http.patch("http://localhost:8080/api/v1/sources/1", () =>
        HttpResponse.json({ error: "toggle failed" }, { status: 500 }),
      ),
    );

    renderWithProviders(<SourceManagementPage />);

    const sourceCard = await screen.findByText("Kubernetes Blog");
    await user.click(within(sourceCard.closest("article")!).getByRole("button", { name: "Enabled" }));
    expect(await screen.findByText("toggle failed")).toBeInTheDocument();
  });

  it("shows an error when the source list cannot be loaded", async () => {
    server.use(
      http.get("http://localhost:8080/api/v1/sources", () =>
        HttpResponse.json({ error: "source list failed" }, { status: 500 }),
      ),
    );

    renderWithProviders(<SourceManagementPage />);

    expect(await screen.findByText("source list failed")).toBeInTheDocument();
  });
});
