/** Formats raw asset_type strings like "collective_investment" → "Collective Investment" */
export function formatAssetType(raw: string): string {
  return raw
    .split('_')
    .map(w => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}
