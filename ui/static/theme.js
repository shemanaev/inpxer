
const themeSwitcher = {
  _scheme: 'auto',
  buttonsTarget: 'button[data-theme-switcher]',
  buttonAttribute: 'data-theme-switcher',
  themeAttribute: 'data-theme-next',
  rootAttribute: 'data-theme',
  localStorageKey: 'preferredColorScheme',

  init() {
    this.scheme = this.getSchemeFromLocalStorage;
    this.initSwitchers();
  },

  get getSchemeFromLocalStorage() {
    return window.localStorage?.getItem(this.localStorageKey) ?? this._scheme;
  },

  initSwitchers() {
    const buttons = document.querySelectorAll(this.buttonsTarget);
    buttons.forEach((button) => {
      if (button.getAttribute(this.buttonAttribute) == this.scheme) {
        button.classList.remove('theme-button-hidden')
      } else {
        button.classList.add('theme-button-hidden')
      }

      button.addEventListener(
        'click',
        (event) => {
          event.preventDefault();
          this.scheme = button.getAttribute(this.themeAttribute);
        },
        false
      );
    });
  },

  set scheme(scheme) {
    if (!['auto', 'light', 'dark'].includes(scheme)) {
      scheme = 'auto'
    }

    this._scheme = scheme;
    this.applyScheme();
    this.saveSchemeToLocalStorage();
    this.setCurrentButton();
  },

  get scheme() {
    return this._scheme;
  },

  applyScheme() {
    let scheme = this._scheme

    if (scheme == 'light' || scheme == 'dark') {
      document.querySelector('html')?.setAttribute(this.rootAttribute, scheme);

    }

    if (scheme == 'auto') {
      document.querySelector('html')?.removeAttribute(this.rootAttribute);
    }
  },

  saveSchemeToLocalStorage() {
    window.localStorage?.setItem(this.localStorageKey, this.scheme);
  },

  setCurrentButton() {
    const buttons = document.querySelectorAll(this.buttonsTarget);
    buttons.forEach((button) => {
      if (button.getAttribute(this.buttonAttribute) == this.scheme) {
        button.classList.remove('theme-button-hidden')
      } else {
        button.classList.add('theme-button-hidden')
      }
    });
  },

  getNextTheme() {
    switch (this.scheme) {
      case 'auto':
        return 'light';
      case 'light':
        return 'dark';
      case 'dark':
        return 'auto';
      default:
        return 'auto';
    }
  },
};

document.addEventListener('DOMContentLoaded', () => {
  themeSwitcher.init();
});
