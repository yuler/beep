json.runs @runs do |run|
  json.partial! "api/v1/beeps/run", run: run
end
json.partial! "api/v1/shared/pagination", page: @page
