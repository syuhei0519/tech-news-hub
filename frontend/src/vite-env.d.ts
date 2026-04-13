/// <reference types="vite/client" />

declare var Worker: {
  prototype: Worker;
  new (scriptURL: string | URL, options?: WorkerOptions): Worker;
};
