export function toSvgDataUri(svg: string): string {
  // Always ensure the SVG is encoded correctly for a data URI.
  // We use encodeURIComponent to handle special characters.
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`
}
