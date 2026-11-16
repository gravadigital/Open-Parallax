import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import Button from '../../ui/button/Button';
import { CloseIcon, MenuIcon, TelescopeLogo } from '../../ui/icons/Icons';
import '../../ui/visually-hidden.css';
import NotificationBell from '../../notifications/notification-bell/NotificationBell';
import { useAuth } from '../../../context/AuthContext';
import { useT } from '../../../i18n';
import LanguageSelect from '../language-select/LanguageSelect';
import UserMenu from '../user-menu/UserMenu';
import './AppHeader.css';

const HOW_IT_WORKS_PATH = '/#como-funciona';

/**
 * Barra global. En desktop: enlaces, accesos e idioma. En mobile: la barra se reduce a logo,
 * campana, avatar y menú, y los enlaces, la cuenta y el idioma viven en el panel desplegable.
 */
const AppHeader: React.FC = () => {
  const { t } = useT();
  const { user, loading } = useAuth();
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);
  const headerRef = React.useRef<HTMLElement>(null);

  const destinations = [
    { id: 'home', to: '/', label: t('nav.home') },
    { id: 'events', to: '/events', label: t('nav.events') },
    { id: 'how-it-works', to: HOW_IT_WORKS_PATH, label: t('nav.howItWorks') },
  ];
  const isActive = (id: string): boolean => {
    if (id === 'home') return location.pathname === '/' && !location.hash;
    if (id === 'events') return location.pathname.startsWith('/events');
    return location.pathname === '/' && location.hash === '#como-funciona';
  };

  const closeMobileMenu = (): void => setMobileMenuOpen(false);

  // El panel es mobile-only: se cierra con Escape, al tocar afuera y al navegar.
  React.useEffect(() => {
    if (!mobileMenuOpen) return undefined;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') closeMobileMenu();
    };
    const onMouseDown = (event: MouseEvent) => {
      if (!headerRef.current?.contains(event.target as Node)) closeMobileMenu();
    };
    document.addEventListener('keydown', onKeyDown);
    document.addEventListener('mousedown', onMouseDown);
    return () => {
      document.removeEventListener('keydown', onKeyDown);
      document.removeEventListener('mousedown', onMouseDown);
    };
  }, [mobileMenuOpen]);

  return (
    <header className="ly-header" ref={headerRef}>
      <div className="ly-header__inner">
        <div className="ly-header__brand">
          <Link to="/" className="ly-header__logo" onClick={closeMobileMenu}>
            <TelescopeLogo className="ly-header__logo-mark" />
            <span className="ui-visually-hidden">{t('common.appName')}</span>
          </Link>
        </div>

        <nav aria-label={t('nav.mainNav')} className="ly-header__links">
          {destinations.map((d) => (
            <Link
              key={d.id}
              to={d.to}
              className={`ly-header__link${isActive(d.id) ? ' ly-header__link--active' : ''}`}
              aria-current={isActive(d.id) ? 'page' : undefined}
            >
              {d.label}
            </Link>
          ))}
        </nav>

        {!loading && (
          <div className="ly-header__actions">
            {user ? (
              <>
                <span data-slot="notifications" className="ly-header__slot--notifications">
                  <NotificationBell />
                </span>
                <UserMenu includeLanguageOptions />
                <LanguageSelect />
              </>
            ) : (
              <>
                <div className="ly-header__auth">
                  <Link to="/login" className="ui-button ui-button--sm ui-button--on-band ly-header__auth-link ly-header__auth-link--login">
                    {t('nav.login')}
                  </Link>
                  <Link to="/register" className="ui-button ui-button--sm ui-button--primary ly-header__auth-link ly-header__auth-link--register">
                    {t('nav.register')}
                  </Link>
                </div>
                <LanguageSelect />
              </>
            )}
          </div>
        )}

        <span className="ly-header__menu-toggle">
          <Button
            variant="icon"
            aria-label={t('nav.openMenu')}
            aria-expanded={mobileMenuOpen}
            aria-controls={mobileMenuOpen ? 'ly-mobile-menu' : undefined}
            className="ly-header__menu-button"
            onClick={() => setMobileMenuOpen((open) => !open)}
          >
            {mobileMenuOpen ? <CloseIcon /> : <MenuIcon />}
          </Button>
        </span>
      </div>

      {mobileMenuOpen && (
        <div id="ly-mobile-menu" className="ly-header__mobile-menu">
          <nav className="ly-header__mobile-nav" aria-label={t('nav.mainNav')}>
            {destinations.map((d) => (
              <Link
                key={d.id}
                to={d.to}
                onClick={closeMobileMenu}
                aria-current={isActive(d.id) ? 'page' : undefined}
                className={`ly-header__mobile-link${isActive(d.id) ? ' ly-header__mobile-link--current' : ''}`}
              >
                {d.label}
              </Link>
            ))}
            {!user && !loading && (
              <Link to="/" onClick={closeMobileMenu} className="ly-header__mobile-link">
                {t('nav.demo')}
              </Link>
            )}
          </nav>

          <div className="ly-header__mobile-divider" />

          {!loading && user && (
            <div className="ly-header__mobile-account">
              <UserMenu includeLanguageOptions />
            </div>
          )}
          {!loading && !user && (
            <div className="ly-header__mobile-auth">
              <Link to="/login" onClick={closeMobileMenu} className="ui-button ui-button--sm ui-button--on-band">
                {t('nav.login')}
              </Link>
              <Link to="/register" onClick={closeMobileMenu} className="ui-button ui-button--sm ui-button--primary">
                {t('nav.register')}
              </Link>
            </div>
          )}

          <div className="ly-header__mobile-divider" />
          <div className="ly-header__mobile-language">
            <span>{t('nav.language')}</span>
            <LanguageSelect />
          </div>
        </div>
      )}
    </header>
  );
};

export default AppHeader;
