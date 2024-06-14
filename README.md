Приложение запускается командой:
```
make up
```
### Схема проекта
![img.png](images_readme/img_8.png)

### Схема БД
![img.png](images_readme/img_9.png)

### Авторизация
#### POST /signin
Результатом успешной авторизации является отдача cookie. Пример запроса: <br/>
![img_2.png](images_readme/img_2.png)
### Регистрация
#### POST /signup
Результатом успешной регистрация является создание нового бользователя в БД. Пример запроса: <br/>
![img_1.png](images_readme/img_1.png)

### Выход
#### DELETE /logout
Для выхода из аккаунта необходима кука session_id, которая была получена при авторизации. <br/>
![img.png](images_readme/img_4.png)

### Проверка авторизации
#### GET /authcheck
Аутентификация пользователя. Проверка происходит по куке session_id. <br/>
![img_3.png](images_readme/img_3.png)

### Вывод списка сотрудников
#### GET /api/v1/employees 
Количество сотрудников настраивается через query параметры. <br/>
![img_5.png](images_readme/img_5.png)

### Подписка на оповещения о дне рожденья сотрудника
#### POST /api/v1/birthday/subscribe
В качестве параметров отправляется айди сотрудника, день рождения которого мы хотим знать. <br/>
![img_6.png](images_readme/img_6.png)

### Отписка от оповещения о дне рожденья сотрудника
#### DELETE /api/v1/birthday/unsubscribe
В качестве параметров отправляется айди сотрудника. <br/>
![img_1.png](images_readme/img_7.png)