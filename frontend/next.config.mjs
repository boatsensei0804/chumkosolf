/** @type {import('next').NextConfig} */
const nextConfig = {
  // standalone output ให้ image prod เล็ก (ตาม CLAUDE.md ข้อ 7)
  output: "standalone",
  reactStrictMode: true,
};

export default nextConfig;
