import { create } from "zustand";

type SearchState = {
  q: string;
  category: string;
  isReadOnly: boolean;
  isFavoriteOnly: boolean;
  setFilters: (next: {
    q: string;
    category: string;
    isReadOnly: boolean;
    isFavoriteOnly: boolean;
  }) => void;
};

export const useSearchStore = create<SearchState>((set) => ({
  q: "",
  category: "",
  isReadOnly: false,
  isFavoriteOnly: false,
  setFilters: (next) => set(next),
}));
