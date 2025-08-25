output "archive_extracted_dir" {
  value = data.temporary_directory.archive.id
}
