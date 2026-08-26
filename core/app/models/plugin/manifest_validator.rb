class Plugin::ManifestValidator
  SUPPORTED_MANIFEST_VERSIONS = [ 1 ].freeze
  SUPPORTED_INPUT_TYPES = %w[ string number boolean url enum secret ].freeze

  attr_reader :errors

  def initialize(manifest)
    @manifest = manifest
    @errors = []
  end

  def valid?
    @errors = []
    validate_structure
    @errors.empty?
  end

  private

  attr_reader :manifest

  def validate_structure
    unless manifest.is_a?(Hash)
      @errors << "must be a JSON object"
      return
    end

    validate_manifest_version
    validate_metadata
    validate_schedule
    validate_inputs
    validate_metrics
    validate_ingest
  end

  def validate_manifest_version
    version = manifest["manifest_version"]
    unless SUPPORTED_MANIFEST_VERSIONS.include?(version)
      @errors << "manifest_version must be one of: #{SUPPORTED_MANIFEST_VERSIONS.join(', ')}"
    end
  end

  def validate_metadata
    %w[ slug name version author ].each do |field|
      value = manifest[field]
      if value.blank? || !value.is_a?(String)
        @errors << "#{field} must be a non-empty string"
      end
    end

    if manifest["slug"].present? && !manifest["slug"].match?(/\A[a-z0-9]+(?:-[a-z0-9]+)*\z/)
      @errors << "slug format must be lowercase kebab-case (e.g. site-uptime)"
    end
  end

  def validate_schedule
    schedule = manifest["schedule"]
    unless schedule.is_a?(Hash)
      @errors << "schedule must be an object"
      return
    end

    default_cron = schedule["default_cron"]
    if default_cron.blank? || !default_cron.is_a?(String)
      @errors << "schedule.default_cron must be a valid cron expression"
    elsif !valid_cron?(default_cron)
      @errors << "schedule.default_cron is not a valid cron expression"
    end

    min_interval = schedule["min_interval_seconds"]
    if min_interval.present? && (!min_interval.is_a?(Integer) || min_interval <= 0)
      @errors << "schedule.min_interval_seconds must be a positive integer"
    end

    failure_threshold = schedule["failure_threshold"]
    if failure_threshold.present? && (!failure_threshold.is_a?(Integer) || failure_threshold < 1)
      @errors << "schedule.failure_threshold must be an integer >= 1"
    end
  end

  def validate_inputs
    inputs = manifest["inputs"]
    return if inputs.nil?

    unless inputs.is_a?(Array)
      @errors << "inputs must be an array"
      return
    end

    seen_names = Set.new
    inputs.each_with_index do |input, index|
      unless input.is_a?(Hash)
        @errors << "inputs[#{index}] must be an object"
        next
      end

      name = input["name"]
      if name.blank? || !name.is_a?(String)
        @errors << "inputs[#{index}].name must be a non-empty string"
      elsif seen_names.include?(name)
        @errors << "inputs[#{index}].name '#{name}' is duplicated"
      else
        seen_names.add(name)
      end

      type = input["type"]
      if type.blank? || !SUPPORTED_INPUT_TYPES.include?(type)
        @errors << "inputs[#{index}].type must be one of: #{SUPPORTED_INPUT_TYPES.join(', ')}"
      end

      if type == "enum"
        options = input["options"]
        unless options.is_a?(Array) && options.all? { |opt| opt.is_a?(String) || opt.is_a?(Numeric) } && options.present?
          @errors << "inputs[#{index}].options must be a non-empty array for enum type"
        end
      end
    end
  end

  def validate_metrics
    metrics = manifest["metrics"]
    return if metrics.nil?

    unless metrics.is_a?(Array)
      @errors << "metrics must be an array"
      return
    end

    seen_names = Set.new
    metrics.each_with_index do |metric, index|
      unless metric.is_a?(Hash)
        @errors << "metrics[#{index}] must be an object"
        next
      end

      name = metric["name"]
      if name.blank? || !name.is_a?(String)
        @errors << "metrics[#{index}].name must be a non-empty string"
      elsif seen_names.include?(name)
        @errors << "metrics[#{index}].name '#{name}' is duplicated"
      else
        seen_names.add(name)
      end

      type = metric["type"]
      unless %w[ number string boolean ].include?(type)
        @errors << "metrics[#{index}].type must be number, string, or boolean"
      end
    end
  end

  def validate_ingest
    ingest = manifest["ingest"]
    return if ingest.nil?

    unless ingest.is_a?(Hash)
      @errors << "ingest must be an object"
      return
    end

    if ingest.key?("webhook") && ![ true, false ].include?(ingest["webhook"])
      @errors << "ingest.webhook must be a boolean"
    end
  end

  def valid_cron?(cron)
    return false if cron.blank?
    parsed = Fugit.parse(cron)
    parsed.is_a?(Fugit::Cron)
  rescue StandardError
    false
  end
end
