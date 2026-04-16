const fs = require('fs');
const langsDir = 'languages';
const publicLangsDir = 'public/languages';

const extraKeys = {
  settings: {
    hideServerStatus: 'Hide server status',
    hideServerStatusHint: 'Hide the server connection status bar',
    hideConsole: 'Hide console',
    hideConsoleHint: 'Hide the console panel'
  },
  profile: {
    subscriptionEnd: 'Valid until',
    noSubscription: 'No active subscription',
    userId: 'User ID',
    memberSince: 'Member since',
    logout: 'Log out'
  },
  auth: {
    login: 'Login',
    register: 'Register',
    loginTitle: 'Login',
    registerTitle: 'Register',
    close: 'Close',
    emailRequired: 'Email is required',
    passwordRequired: 'Password is required',
    usernameRequired: 'Username is required',
    invalidEmail: 'Invalid email format',
    passwordMinLength: 'Password must be at least 6 characters',
    passwordsNotMatch: 'Passwords do not match',
    password: 'Password',
    enterPassword: 'Enter password',
    loggingIn: 'Logging in...',
    username: 'Username',
    yourName: 'Your name',
    min6Chars: 'Min. 6 characters',
    confirmPassword: 'Confirm password',
    repeatPassword: 'Repeat password',
    registering: 'Registering...',
    noAccount: "Don't have an account?",
    hasAccount: 'Already have an account?'
  }
};

const translations = {
  ru: {
    settings: { hideServerStatus: 'Скрыть статус сервера', hideServerStatusHint: 'Скрыть строку состояния подключения сервера', hideConsole: 'Скрыть консоль', hideConsoleHint: 'Скрыть панель консоли' },
    profile: { subscriptionEnd: 'Действительно до', noSubscription: 'Нет активной подписки', userId: 'ID пользователя', memberSince: 'Участник с', logout: 'Выйти' },
    auth: { login: 'Войти', register: 'Регистрация', loginTitle: 'Войти', registerTitle: 'Регистрация', close: 'Закрыть', emailRequired: 'Email обязателен', passwordRequired: 'Пароль обязателен', usernameRequired: 'Имя пользователя обязательно', invalidEmail: 'Неверный формат email', passwordMinLength: 'Пароль должен содержать минимум 6 символов', passwordsNotMatch: 'Пароли не совпадают', password: 'Пароль', enterPassword: 'Введите пароль', loggingIn: 'Вход...', username: 'Имя пользователя', yourName: 'Ваше имя', min6Chars: 'Мин. 6 символов', confirmPassword: 'Подтвердите пароль', repeatPassword: 'Повторите пароль', registering: 'Регистрация...', noAccount: 'Нет аккаунта?', hasAccount: 'Уже есть аккаунт?' }
  },
  ko: {
    settings: { hideServerStatus: '서버 상태 숨기기', hideServerStatusHint: '서버 연결 상태 표시줄 숨기기', hideConsole: '콘솔 숨기기', hideConsoleHint: '콘솔 패널 숨기기' },
    profile: { subscriptionEnd: '유효 기간', noSubscription: '활성 구독 없음', userId: '사용자 ID', memberSince: '등록일', logout: '로그아웃' },
    auth: { login: '로그인', register: '회원가입', loginTitle: '로그인', registerTitle: '회원가입', close: '닫기', emailRequired: '이메일은 필수입니다', passwordRequired: '비밀번호는 필수입니다', usernameRequired: '사용자 이름은 필수입니다', invalidEmail: '이메일 형식이 잘못되었습니다', passwordMinLength: '비밀번호는 최소 6자 이상이어야 합니다', passwordsNotMatch: '비밀번호가 일치하지 않습니다', password: '비밀번호', enterPassword: '비밀번호 입력', loggingIn: '로그인 중...', username: '사용자 이름', yourName: '이름', min6Chars: '최소 6자', confirmPassword: '비밀번호 확인', repeatPassword: '비밀번호 재입력', registering: '가입 중...', noAccount: '계정이 없으신가요?', hasAccount: '이미 계정이 있으신가요?' }
  },
  cs: {
    settings: { hideServerStatus: 'Skrýt stav serveru', hideServerStatusHint: 'Skrýt stavový řádek připojení serveru', hideConsole: 'Skrýt konzoli', hideConsoleHint: 'Skrýt panel konzole' },
    profile: { subscriptionEnd: 'Platné do', noSubscription: 'Žádné aktivní předplatné', userId: 'ID uživatele', memberSince: 'Členem od', logout: 'Odhlásit' },
    auth: { login: 'Přihlásit', register: 'Registrace', loginTitle: 'Přihlášení', registerTitle: 'Registrace', close: 'Zavřít', emailRequired: 'Email je povinný', passwordRequired: 'Heslo je povinné', usernameRequired: 'Uživatelské jméno je povinné', invalidEmail: 'Neplatný formát emailu', passwordMinLength: 'Heslo musí mít alespoň 6 znaků', passwordsNotMatch: 'Hesla se neshodují', password: 'Heslo', enterPassword: 'Zadejte heslo', loggingIn: 'Přihlašování...', username: 'Uživatelské jméno', yourName: 'Vaše jméno', min6Chars: 'Min. 6 znaků', confirmPassword: 'Potvrďte heslo', repeatPassword: 'Zopakujte heslo', registering: 'Registrace...', noAccount: 'Nemáte účet?', hasAccount: 'Už máte účet?' }
  },
  sk: {
    settings: { hideServerStatus: 'Skryť stav servera', hideServerStatusHint: 'Skryť stavový riadok pripojenia servera', hideConsole: 'Skryť konzolu', hideConsoleHint: 'Skryť panel konzoly' },
    profile: { subscriptionEnd: 'Platné do', noSubscription: 'Žiadne aktívne predplatné', userId: 'ID používateľa', memberSince: 'Členom od', logout: 'Odhlásiť' },
    auth: { login: 'Prihlásiť', register: 'Registrácia', loginTitle: 'Prihlásenie', registerTitle: 'Registrácia', close: 'Zatvoriť', emailRequired: 'Email je povinný', passwordRequired: 'Heslo je povinné', usernameRequired: 'Používateľské meno je povinné', invalidEmail: 'Neplatný formát emailu', passwordMinLength: 'Heslo musí mať aspoň 6 znakov', passwordsNotMatch: 'Heslá sa nezhodujú', password: 'Heslo', enterPassword: 'Zadajte heslo', loggingIn: 'Prihlasovanie...', username: 'Používateľské meno', yourName: 'Vaše meno', min6Chars: 'Min. 6 znakov', confirmPassword: 'Potvrďte heslo', repeatPassword: 'Zopakujte heslo', registering: 'Registrácia...', noAccount: 'Nemáte účet?', hasAccount: 'Už máte účet?' }
  },
  sl: {
    settings: { hideServerStatus: 'Skrij status strežnika', hideServerStatusHint: 'Skrij vrstico stanja povezave strežnika', hideConsole: 'Skrij konzolo', hideConsoleHint: 'Skrij ploščo konzole' },
    profile: { subscriptionEnd: 'Veljavno do', noSubscription: 'Brez aktivne naročnine', userId: 'ID uporabnika', memberSince: 'Član od', logout: 'Odjava' },
    auth: { login: 'Prijava', register: 'Registracija', loginTitle: 'Prijava', registerTitle: 'Registracija', close: 'Zapri', emailRequired: 'Email je obvezen', passwordRequired: 'Geslo je obvezno', usernameRequired: 'Uporabniško ime je obvezno', invalidEmail: 'Neveljaven format emaila', passwordMinLength: 'Geslo mora imeti vsaj 6 znakov', passwordsNotMatch: 'Gesli se ne ujemata', password: 'Geslo', enterPassword: 'Vnesite geslo', loggingIn: 'Prijava...', username: 'Uporabniško ime', yourName: 'Vaše ime', min6Chars: 'Min. 6 znakov', confirmPassword: 'Potrdite geslo', repeatPassword: 'Ponovite geslo', registering: 'Registracija...', noAccount: 'Nimate računa?', hasAccount: 'Že imate račun?' }
  },
  hr: {
    settings: { hideServerStatus: 'Sakrij status poslužitelja', hideServerStatusHint: 'Sakrij traku statusa veze poslužitelja', hideConsole: 'Sakrij konzolu', hideConsoleHint: 'Sakrij ploču konzole' },
    profile: { subscriptionEnd: 'Vrijedi do', noSubscription: 'Nema aktivne pretplate', userId: 'Korisnički ID', memberSince: 'Član od', logout: 'Odjava' },
    auth: { login: 'Prijava', register: 'Registracija', loginTitle: 'Prijava', registerTitle: 'Registracija', close: 'Zatvori', emailRequired: 'Email je obavezan', passwordRequired: 'Lozinka je obavezna', usernameRequired: 'Korisničko ime je obavezno', invalidEmail: 'Neispravan format emaila', passwordMinLength: 'Lozinka mora imati barem 6 znakova', passwordsNotMatch: 'Lozinke se ne podudaraju', password: 'Lozinka', enterPassword: 'Unesite lozinku', loggingIn: 'Prijava...', username: 'Korisničko ime', yourName: 'Vaše ime', min6Chars: 'Min. 6 znakova', confirmPassword: 'Potvrdite lozinku', repeatPassword: 'Ponovite lozinku', registering: 'Registracija...', noAccount: 'Nemate račun?', hasAccount: 'Već imate račun?' }
  },
  sr: {
    settings: { hideServerStatus: 'Сакриј статус сервера', hideServerStatusHint: 'Сакриј траку статуса везе сервера', hideConsole: 'Сакриј конзолу', hideConsoleHint: 'Сакриј панел конзоле' },
    profile: { subscriptionEnd: 'Важи до', noSubscription: 'Нема активне претплате', userId: 'Кориснички ID', memberSince: 'Члан од', logout: 'Одјава' },
    auth: { login: 'Пријава', register: 'Регистрација', loginTitle: 'Пријава', registerTitle: 'Регистрација', close: 'Затвори', emailRequired: 'Email је обавезан', passwordRequired: 'Лозинка је обавезна', usernameRequired: 'Корисничко име је обавезно', invalidEmail: 'Неисправан формат email-а', passwordMinLength: 'Лозинка мора имати најмање 6 карактера', passwordsNotMatch: 'Лозинке се не подударају', password: 'Лозинка', enterPassword: 'Унесите лозинку', loggingIn: 'Пријава...', username: 'Корисничко име', yourName: 'Ваше име', min6Chars: 'Мин. 6 карактера', confirmPassword: 'Потврдите лозинку', repeatPassword: 'Поновите лозинку', registering: 'Регистрација...', noAccount: 'Немате налог?', hasAccount: 'Већ имате налог?' }
  },
  bg: {
    settings: { hideServerStatus: 'Скрий статус сървъра', hideServerStatusHint: 'Скрий лентата за състояние на сървъра', hideConsole: 'Скрий конзолата', hideConsoleHint: 'Скрий панела на конзолата' },
    profile: { subscriptionEnd: 'Валидно до', noSubscription: 'Няма активна абонамент', userId: 'Потребител ID', memberSince: 'Член от', logout: 'Изход' },
    auth: { login: 'Вход', register: 'Регистрация', loginTitle: 'Вход', registerTitle: 'Регистрация', close: 'Затвори', emailRequired: 'Email е задължителен', passwordRequired: 'Паролата е задължителна', usernameRequired: 'Потребителското име е задължително', invalidEmail: 'Невалиден формат на email', passwordMinLength: 'Паролата трябва да е поне 6 символа', passwordsNotMatch: 'Паролите не съвпадат', password: 'Парола', enterPassword: 'Въведете парола', loggingIn: 'Влизане...', username: 'Потребителско име', yourName: 'Вашето име', min6Chars: 'Мин. 6 символа', confirmPassword: 'Потвърдете паролата', repeatPassword: 'Повторете паролата', registering: 'Регистрация...', noAccount: 'Нямате акаунт?', hasAccount: 'Вече имате акаунт?' }
  },
  ro: {
    settings: { hideServerStatus: 'Ascunde status server', hideServerStatusHint: 'Ascunde bara de stare a conexiunii serverului', hideConsole: 'Ascunde consola', hideConsoleHint: 'Ascunde panoul consolei' },
    profile: { subscriptionEnd: 'Valabil până la', noSubscription: 'Niciun abonament activ', userId: 'ID utilizator', memberSince: 'Membru din', logout: 'Deconectare' },
    auth: { login: 'Autentificare', register: 'Înregistrare', loginTitle: 'Autentificare', registerTitle: 'Înregistrare', close: 'Închide', emailRequired: 'Emailul este obligatoriu', passwordRequired: 'Parola este obligatorie', usernameRequired: 'Numele de utilizator este obligatoriu', invalidEmail: 'Format email invalid', passwordMinLength: 'Parola trebuie să aibă cel puțin 6 caractere', passwordsNotMatch: 'Parolele nu se potrivesc', password: 'Parolă', enterPassword: 'Introduceți parola', loggingIn: 'Autentificare...', username: 'Nume utilizator', yourName: 'Numele dvs', min6Chars: 'Min. 6 caractere', confirmPassword: 'Confirmați parola', repeatPassword: 'Repetați parola', registering: 'Înregistrare...', noAccount: 'Nu aveți cont?', hasAccount: 'Aveți deja cont?' }
  }
};

// First copy all files from public/languages to languages
const files = fs.readdirSync(publicLangsDir);
files.forEach(f => {
  if (!f.endsWith('.json')) return;
  const srcPath = publicLangsDir + '/' + f;
  const destPath = langsDir + '/' + f;
  const data = JSON.parse(fs.readFileSync(srcPath, 'utf8'));
  
  // Apply language-specific overrides
  const lang = f.replace('.json', '');
  if (translations[lang]) {
    Object.assign(data.settings, translations[lang].settings);
    Object.assign(data.profile, translations[lang].profile);
    data.auth = translations[lang].auth;
  }
  
  fs.writeFileSync(destPath, JSON.stringify(data, null, 2), 'utf8');
  console.log('Updated ' + f);
});
console.log('Done!');
