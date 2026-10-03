"""Regression: a resource and data source sharing a name need separate references."""
import unittest
from schema_reference import inventory, reference_name


class SharedTypeNames(unittest.TestCase):
    def test_collision_keeps_both_schemas_and_existing_urls(self):
        name = 'prisma-airs_supply_chain_skill_scanning_instance'
        resources = {name}
        self.assertEqual(reference_name(name, 'resource', resources), name)
        self.assertEqual(reference_name(name, 'data source', resources), 'data-source-' + name)
        self.assertEqual(reference_name('unique', 'data source', resources), 'unique')
        catalog = [{'label': 'Supply Chain', 'implemented': True,
                    'resources': [{'name': name, 'guide': 'resources/instance'}],
                    'data_sources': [{'name': name, 'guide': 'data-sources/instance'}]}]
        page = inventory(catalog)
        self.assertIn('generated/' + name + '.md', page)
        self.assertIn('generated/data-source-' + name + '.md', page)


if __name__ == '__main__':
    unittest.main()
