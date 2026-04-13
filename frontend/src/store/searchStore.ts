import { create } from "zustand";

type SearchState = {
  q: string;
  category: string;
  setFilters: (next: { q: string; category: string }) => void;
};

export const useSearchStore = create<SearchState>((set) => ({
  q: "",
  category: "",
  setFilters: (next) => set(next),
}));
