import octraMascot from '../../../images/octra-mascot.png';

const links = [
  { href: '#showcase', label: 'Showcase' },
  { href: '#audiences', label: 'Audiences' },
  { href: '#providers', label: 'Models' },
  { href: '/app', label: 'Open app' },
];

export function FooterSection() {
  return (
    <footer className="border-t border-white/10 bg-black px-4 py-10 text-white sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-7xl flex-col gap-8 md:flex-row md:items-center md:justify-between">
        <div className="max-w-md">
          <div className="mb-4 flex items-center gap-3">
            <img src={octraMascot} alt="Octra Mascot" className="h-9 w-9 rounded-md object-contain" />
            <span className="text-lg font-semibold">Octra</span>
          </div>
          <p className="text-sm leading-6 text-white/58">
            AI task execution for code, research, text documents, and presentations.
          </p>
        </div>

        <nav className="flex flex-wrap gap-x-6 gap-y-3">
          {links.map((link) => (
            <a key={link.href} href={link.href} className="text-sm font-medium text-white/58 transition-colors hover:text-white">
              {link.label}
            </a>
          ))}
        </nav>
      </div>
    </footer>
  );
}
