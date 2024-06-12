package pkg

var BirthdaySub = "INSERT INTO subscriber (id_subscribe_from, id_subscribe_to) VALUES ($1, $2)"
var BirthdayUnSub = "DELETE FROM subscriber WHERE id_subscribe_from = $1 AND id_subscribe_to = $2"

var GetUser = "SELECT profile.id, profile.login, profile.email, profile.birthday FROM profile WHERE profile.login = $1 AND profile.password = $2"
var FindUser = "SELECT login FROM profile WHERE login = $1"
var CreateUser = "INSERT INTO profile(login, password, email, birthday) VALUES($1, $2, $3, $4) RETURNING id"
var GetUserId = "SELECT profile.id FROM profile WHERE profile.login = $1"
var GetEmployees = "SELECT profile.id, profile.login, profile.email, profile.birthday FROM profile OFFSET $1 LIMIT $2"
