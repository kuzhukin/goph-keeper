# goph-keeper

Схема
1) Регистрация. Юзер вводит команду client register --login LOGIN --pasword PASSWORD. данные аплоадятся на сервер и в случае успеха записываются в БД
2) Юзер создает файл командой client new -f FILE (приложение создает новую запись в БД с пустым содержимым)
3) Юзер редактирует файл командой client edit -f FILE (приложение находит файл в БД, создает временный файл, запускает текстовый редактор, если файл изменен - аплоадит его в БД)
4) Юзер аплоадит файл командой client push -f FILE (приложение достает файл, аплоадит его на сервер и если успешно апает ревизию)


Вспомогательные 
client config -s SERVER_URL
