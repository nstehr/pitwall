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

ActiveRecord::Schema[7.0].define(version: 2022_07_30_025057) do
  create_table "orchestrators", force: :cascade do |t|
    t.string "status"
    t.string "name"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "virtual_machines_count"
    t.index ["name"], name: "index_orchestrators_on_name", unique: true
  end

  create_table "virtual_machines", force: :cascade do |t|
    t.string "image"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "orchestrator_id"
    t.string "status"
    t.index ["orchestrator_id"], name: "index_virtual_machines_on_orchestrator_id"
  end

  add_foreign_key "virtual_machines", "orchestrators"
end
