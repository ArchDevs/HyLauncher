import type { Translations } from "../types";

export const ru: Translations = {
  common: {
    play: "ИГРАТЬ",
    install: "УСТАНОВКА...",
    ready: "Готово",
    cancel: "Отмена",
    close: "Закрыть",
    delete: "Удалить",
    confirm: "Подтвердить",
    update: "Обновить",
    updateAvailable: "Доступно обновление",
    updating: "Обновление",
    error: "Ошибка",
    copy: "Копировать",
    copied: "Скопировано!",
  },
  pages: {
    home: "Главная",
    servers: "Серверы",
  },
  profile: {
    username: "Имя пользователя",
    version: "Версия",
    noVersion: "Нет",
    releaseType: {
      preRelease: "Pre-Release",
      release: "Release",
    },
  },
  control: {
    status: {
      readyToPlay: "Готов к игре",
    },
    updateAvailable: "Доступно обновление",
  },
  modals: {
    delete: {
      title: "Вы уверены?",
      message: "Вы точно хотите удалить игру?",
      warning: "Это действие удалит все файлы игры без возможности восстановления!",
      confirmButton: "Удалить всё",
      cancelButton: "Отмена",
    },
    error: {
      title: "Произошла ошибка",
      technicalDetails: "Технические детали",
      stackTrace: "Трассировка стека",
      suggestions: {
        network: "Проверьте подключение к интернету и попробуйте снова.",
        filesystem: "Убедитесь, что у вас достаточно места на диске и у лаунчера есть необходимые права доступа.",
        validation: "Пожалуйста, проверьте введенные данные и попробуйте снова.",
        game: "Попробуйте перезапустить лаунчер или переустановить игру.",
        default: "Пожалуйста, сообщите об этой проблеме, если она сохраняется.",
      },
    },
    update: {
      title: "ОБНОВЛЕНИЕ ЛАУНЧЕРА",
      message: "Загрузка последней версии. HyLauncher автоматически перезапустится после завершения.",
    },
  },
  banners: {
    hynexus: {
      text: "HyNexus - это Hytale, каким он должен быть. Экономика, Кланы, PVP, PVE, ждем тебя!",
    },
  },
};

