# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[7.0].define(version: 2023_06_07_011623) do
  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"

  create_table "identities", force: :cascade do |t|
    t.text "public_key"
    t.text "ziti_identity"
    t.bigint "user_id", null: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["user_id"], name: "index_identities_on_user_id"
  end

  create_table "orchestrators", force: :cascade do |t|
    t.string "status"
    t.string "name"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "virtual_machines_count"
    t.string "health_check_url"
    t.index ["name"], name: "index_orchestrators_on_name", unique: true
  end

  create_table "services", force: :cascade do |t|
    t.text "name"
    t.integer "port"
    t.boolean "private"
    t.string "protocol"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.bigint "virtual_machine_id", null: false
    t.index ["virtual_machine_id"], name: "index_services_on_virtual_machine_id"
  end

  create_table "users", force: :cascade do |t|
    t.string "email", default: "", null: false
    t.string "encrypted_password", default: "", null: false
    t.string "reset_password_token"
    t.datetime "reset_password_sent_at"
    t.datetime "remember_created_at"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.string "provider"
    t.string "uid"
    t.integer "virtual_machines_count"
    t.string "username"
    t.string "name"
    t.index ["email"], name: "index_users_on_email", unique: true
    t.index ["reset_password_token"], name: "index_users_on_reset_password_token", unique: true
  end

  create_table "virtual_machines", force: :cascade do |t|
    t.string "image"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.bigint "orchestrator_id"
    t.string "status"
    t.string "public_key"
    t.bigint "user_id", null: false
    t.string "name"
    t.index ["name", "user_id"], name: "index_virtual_machines_on_name_and_user_id", unique: true
    t.index ["orchestrator_id"], name: "index_virtual_machines_on_orchestrator_id"
    t.index ["user_id"], name: "index_virtual_machines_on_user_id"
  end

  add_foreign_key "identities", "users"
  add_foreign_key "services", "virtual_machines"
  add_foreign_key "virtual_machines", "orchestrators"
  add_foreign_key "virtual_machines", "users"
end
