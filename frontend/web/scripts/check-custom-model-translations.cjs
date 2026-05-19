const translate = require('@vitalets/google-translate-api').translate;
const fs = require('fs');
const path = require('path');

const SOURCE_LANG = 'en';
const TARGET_DIRS = [
  'languages',
  'public/languages',
];

async function translateText(text, targetLang) {
  if (!text || typeof text !== 'string') return text;
  try {
    const res = await translate(text, { from: SOURCE_LANG, to: targetLang });
    return res.text;
  } catch (e) {
    console.warn(`  [warn] Failed "${text.substring(0, 40)}..." → ${targetLang}: ${e.message}`);
    return text;
  }
}

async function translateObject(obj, targetLang) {
  const result = {};
  for (const key of Object.keys(obj)) {
    const value = obj[key];
    if (typeof value === 'string') {
      result[key] = await translateText(value, targetLang);
    } else if (value && typeof value === 'object' && !Array.isArray(value)) {
      result[key] = await translateObject(value, targetLang);
    } else {
      result[key] = value;
    }
    await new Promise(r => setTimeout(r, 120)); // polite delay
  }
  return result;
}

async function main() {
  const enPath = path.join('languages', 'en.json');
  const enData = JSON.parse(fs.readFileSync(enPath, 'utf8'));

  if (!enData.landing) {
    console.error('No "landing" section in en.json');
    process.exit(1);
  }

  const landingEn = enData.landing;

  for (const dir of TARGET_DIRS) {
    const files = fs.readdirSync(dir).filter(f => f.endsWith('.json') && f !== 'en.json');

    for (const file of files) {
      const langCode = file.replace('.json', '');
      const filePath = path.join(dir, file);

      console.log(`\n→ Translating landing for ${langCode} (${dir})`);

      const data = JSON.parse(fs.readFileSync(filePath, 'utf8'));
      const translated = await translateObject(landingEn, langCode);

      data.landing = translated;
      fs.writeFileSync(filePath, JSON.stringify(data, null, 2));
      console.log(`   ✓ Saved ${file}`);
    }
  }

  console.log('\n✅ Finished translating landing section for all languages');
}

main().catch(console.error);