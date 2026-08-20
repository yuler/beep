json.intent @proposal.intent
json.title @proposal.title
json.body @proposal.body
json.run_at @proposal.run_at&.iso8601
json.timezone @proposal.timezone
json.errors @proposal.errors
json.confirmable @proposal.confirmable?
json.message @proposal.message
