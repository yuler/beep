module VersionHelper
  def app_version_badge
    tag.span(Rails.application.config.app_version, class: "app-version-badge")
  end

  def git_revision_badge
    tag.span(Rails.application.config.git_revision.slice(0, 7), class: "git-revision-badge")
  end
end
