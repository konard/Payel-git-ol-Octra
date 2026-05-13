const fs = require('fs');
const path = require('path');

const root = path.resolve(__dirname, '..');
const languageDirs = [
  path.join(root, 'languages'),
  path.join(root, 'public', 'languages'),
];

const requiredKeys = [
  'settings.customModels',
  'models.title',
  'models.description',
  'models.addNew',
  'models.name',
  'models.namePlaceholder',
  'models.provider',
  'models.noProvider',
  'models.save',
  'models.cancel',
  'models.delete',
  'models.confirmDelete',
];

function getNestedValue(data, keyPath) {
  return keyPath.split('.').reduce((current, key) => {
    if (current && typeof current === 'object' && key in current) {
      return current[key];
    }
    return undefined;
  }, data);
}

const failures = [];

for (const dir of languageDirs) {
  const files = fs.readdirSync(dir).filter((file) => file.endsWith('.json')).sort();

  for (const file of files) {
    const fullPath = path.join(dir, file);
    const data = JSON.parse(fs.readFileSync(fullPath, 'utf8'));

    for (const key of requiredKeys) {
      const value = getNestedValue(data, key);
      if (typeof value !== 'string' || value.trim() === '' || value === key) {
        failures.push(`${path.relative(root, fullPath)}: missing ${key}`);
      }
    }
  }
}

if (failures.length > 0) {
  console.error('Custom model translation validation failed:');
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exit(1);
}

console.log('Custom model translations are complete.');
