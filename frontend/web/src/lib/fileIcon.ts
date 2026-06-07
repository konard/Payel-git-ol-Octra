import iconData from './fileIconData.json';

const { fileExtensions, fileNames } = iconData as {
  fileExtensions: Record<string, string>;
  fileNames: Record<string, string>;
};

export function getFileIconName(filePath: string): string {
  const parts = filePath.replace(/\\/g, '/').split('/');
  const name = parts.pop() || '';

  const fullName = name;
  if (fileNames[fullName]) return fileNames[fullName];

  const dotIndex = name.lastIndexOf('.');
  if (dotIndex > 0) {
    const ext = name.slice(dotIndex).toLowerCase();
    if (fileExtensions[ext]) return fileExtensions[ext];
  }

  for (let i = name.lastIndexOf('.') - 1; i > 0; i = name.lastIndexOf('.', i - 1)) {
    const ext = name.slice(i).toLowerCase();
    if (fileExtensions[ext]) return fileExtensions[ext];
  }

  return 'file';
}
