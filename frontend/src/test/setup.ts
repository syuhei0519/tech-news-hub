import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach, beforeAll } from "vitest";
import { setupServer } from "msw/node";
import { api } from "../lib/api";
import { useSearchStore } from "../store/searchStore";

export const server = setupServer();

beforeAll(() => {
  api.defaults.baseURL = "http://localhost:8080";
  server.listen({ onUnhandledRequest: "error" });
});

afterEach(() => {
  cleanup();
  server.resetHandlers();
  useSearchStore.setState(useSearchStore.getInitialState());
});

afterAll(() => {
  server.close();
});
