import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

// ล้าง DOM หลังแต่ละ test เพื่อกัน state รั่วข้าม test
afterEach(() => {
  cleanup();
});

// antd บางคอมโพเนนต์เรียก matchMedia — jsdom ไม่มี ต้อง stub
if (typeof window !== "undefined" && !window.matchMedia) {
  window.matchMedia = (query: string): MediaQueryList =>
    ({
      matches: false,
      media: query,
      onchange: null,
      addListener: () => {},
      removeListener: () => {},
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false,
    }) as MediaQueryList;
}

// antd Drawer/Modal วัด scrollbar ด้วย getComputedStyle(el, pseudoElt) ซึ่ง jsdom ยังไม่รองรับ
// ตัด argument ที่สองออกเพื่อกัน warning รก ๆ ใน test (ไม่กระทบ logic)
if (typeof window !== "undefined") {
  const original = window.getComputedStyle.bind(window);
  window.getComputedStyle = ((el: Element) => original(el)) as typeof window.getComputedStyle;
}
