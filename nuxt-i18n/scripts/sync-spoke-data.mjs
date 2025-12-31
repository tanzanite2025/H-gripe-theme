import fs from 'fs/promises';
import path from 'path';
import { fileURLToPath } from 'url';

// Config
const API_URL = process.env.WP_API_URL || 'https://tanzanite.site/wp-json/tanzanite/v1/spoke-db-export';
const TARGET_FILE = path.join(path.dirname(fileURLToPath(import.meta.url)), '../app/data/spoke-calculator/database.ts');

async function main() {
    console.log(`🔌 Connecting to ${API_URL}...`);

    try {
        const response = await fetch(API_URL);
        if (!response.ok) {
            throw new Error(`API responded with ${response.status}: ${response.statusText}`);
        }
        const data = await response.json();
        console.log(`📦 Data received: ${data.rims.length} rim brands, ${data.hubs.length} hub brands.`);

        // 1. Generate Types & Interfaces (Static)
        // We keep the interfaces hardcoded in the script or preserve them from file?
        // It's safer to generate them to ensure they match the data structure we verify.
        // BUT, PRESET_BUILDS at the bottom relies on them.
        // Let's generate the top part of the file.

        const fileHeader = `// AUTO-GENERATED FILE. DO NOT EDIT BRANDS MANUALLY.
// Run "npm run sync-data" to update from WordPress.

export interface SpokeGeometry {
  leftFlange: number
  rightFlange: number
  leftFlangePcd: number
  rightFlangePcd: number
}

export interface RimModel {
  id: string
  name: string
  erd: number
}

export interface HubModel {
  id: string
  name: string
  front?: SpokeGeometry
  rear?: SpokeGeometry
}

export interface Brand<T> {
  id: string
  name: string
  items: T[]
}

export const RIM_DATABASE: Brand<RimModel>[] = ${JSON.stringify(data.rims, null, 4)};

export const HUB_DATABASE: Brand<HubModel>[] = ${JSON.stringify(data.hubs, null, 4)};
`;

        // 2. Read existing file to preserve PRESET_BUILDS
        let existingContent = '';
        try {
            existingContent = await fs.readFile(TARGET_FILE, 'utf-8');
        } catch (e) {
            console.warn('⚠️ Could not read existing file. PRESET_BUILDS might be lost if not handled.');
        }

        // Extract PRESET_BUILDS section
        const presetMarker = 'export interface WheelBuildPreset';
        const splitParts = existingContent.split(presetMarker);

        let preserData = '';
        if (splitParts.length > 1) {
            preserData = presetMarker + splitParts[1];
        } else {
            // Fallback if marker not found (e.g. first run or file corrupted), verify if we need to restore default presets
            console.warn('⚠️ PRESET_BUILDS section not found in existing file. Appending default empty structure.');
            preserData = `
export interface WheelBuildPreset {
  id: string
  name: string
  keywords: string[]
  description?: string
  
  // Configuration
  rimBrandId: string
  rimModelId: string
  hubBrandId: string
  hubModelId: string
  spokeCount: number
  crossing: number
  nippleType: 'standard' | 'hidden'
  nippleLength: number | null
}

export const PRESET_BUILDS: WheelBuildPreset[] = [];
`;
        }

        // 3. Combine and Write
        const finalContent = fileHeader + '\n\n' + preserData;
        await fs.writeFile(TARGET_FILE, finalContent, 'utf-8');

        console.log(`✅ Database synced successfully to ${TARGET_FILE}`);

    } catch (err) {
        console.error('❌ Sync failed:', err.message);
        process.exit(1);
    }
}

main();
