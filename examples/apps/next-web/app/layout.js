export const metadata = {
  title: "ELC Next example",
  description: "Next.js app running with hot reload via elc",
};

export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
