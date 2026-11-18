import React from 'react';
import { Link } from 'react-router-dom';
import { useT } from '../../../i18n';
import { GitHubIcon, OpenParallaxLogo } from '../../ui/icons/Icons';
import '../../ui/visually-hidden.css';
import './AppFooter.css';

/** Pie global: comparte la banda negra del header; marca, versión, repositorio y contacto. */
const AppFooter: React.FC = () => {
  const { t } = useT();

  return (
    <footer className="ly-footer">
      <div className="ly-footer__inner">
        <div className="ly-footer__brand">
          <Link to="/" className="ly-footer__logo">
            <OpenParallaxLogo className="ly-footer__logo-mark" />
            <span className="ui-visually-hidden">{t('common.appName')}</span>
          </Link>
          <span className="ly-footer__version">{t('footer.version')}</span>
        </div>

        <a
          className="ly-footer__github"
          href="https://github.com/gravadigital/openparallax"
          target="_blank"
          rel="noopener noreferrer"
        >
          <GitHubIcon className="ly-footer__github-icon" />
          <span className="ly-footer__github-text">
            <span>{t('footer.github')}</span>
            <span className="ly-footer__repo">{t('footer.repo')}</span>
          </span>
        </a>

        <div className="ly-footer__contact">
          <span className="ly-footer__eyebrow">{t('footer.contact')}</span>
          <a href="mailto:info@grava.digital">{t('footer.email')}</a>
          <a href="tel:+5491130942020">{t('footer.phone')}</a>
        </div>
      </div>
    </footer>
  );
};

export default AppFooter;
