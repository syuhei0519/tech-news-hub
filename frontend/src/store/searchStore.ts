import { create } from "zustand";

type SearchState = {
  q: string;
  category: string;
  sourceId: string;
  isReadOnly: boolean;
  isFavoriteOnly: boolean;
  from: string;
  to: string;
  setFilters: (next: {
    q: string;
    category: string;
    sourceId: string;
    isReadOnly: boolean;
    isFavoriteOnly: boolean;
    from: string;
    to: string;
  }) => void;
};

export const useSearchStore = create<SearchState>((set) => ({
  q: "",
  category: "",
  sourceId: "",
  isReadOnly: false,
  isFavoriteOnly: false,
  from: "",
  to: "",
  setFilters: (next) => set(next),
}));
