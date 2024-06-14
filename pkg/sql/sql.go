package pkg

var BirthdaySub = "INSERT INTO subscriber (id_subscribe_from, id_subscribe_to) VALUES ($1, $2)"
var BirthdayUnSub = "DELETE FROM subscriber WHERE id_subscribe_from = $1 AND id_subscribe_to = $2"

var GetUser = "SELECT profile.id, profile.login, profile.email, profile.birthday FROM profile WHERE profile.login = $1 AND profile.password = $2"
var FindUser = "SELECT login FROM profile WHERE login = $1"
var CreateUser = "INSERT INTO profile(login, password, email, birthday) VALUES($1, $2, $3, $4) RETURNING id"
var GetUserId = "SELECT profile.id FROM profile WHERE profile.login = $1"
var GetEmployees = "SELECT profile.id, profile.login, profile.email, profile.birthday FROM profile OFFSET $1 LIMIT $2"
var GetBirthdayEmployees = `SELECT id, login, email, birthday FROM profile WHERE EXTRACT(DAY FROM birthday) = EXTRACT(DAY FROM CURRENT_DATE) AND EXTRACT(MONTH FROM birthday) = EXTRACT(MONTH FROM CURRENT_DATE)`
var GetEmployeeByBirthday = "SELECT p.id, p.login, p.email, p.birthday FROM profile p JOIN subscriber s ON p.id = s.id_subscribe_from WHERE s.id_subscribe_to = $1"
var GetEmployeesBySubId = "SELECT users.id, user.login, user.email, user.birthday FROM users WHERE id_subscribe_to=$1"
